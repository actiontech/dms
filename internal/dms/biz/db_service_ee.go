//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	v1Base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	v1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	pkgPeriods "github.com/actiontech/dms/pkg/periods"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

const ImportDBServicesCsvColumnsNum = 14

type RuleTemplate struct {
	RuleTemplateName string `json:"rule_template_name"`
	RuleTemplateID   string `json:"rule_template_id"`
	DbType           string `json:"db_type"`
}

type ImportDbServicesPreCheckErr string

func (e ImportDbServicesPreCheckErr) Error() string {
	return string(e)
}

var (
	IDBPCErrInvalidInput             ImportDbServicesPreCheckErr = "若无特别说明参数均为必填"
	IDBPCErrProjNonExist             ImportDbServicesPreCheckErr = "所属项目不存在"
	IDBPCErrProjNotActive            ImportDbServicesPreCheckErr = "所属项目状态异常"
	IDBPCErrProjNotAllowed           ImportDbServicesPreCheckErr = "所属项目不是操作中的项目"
	IDBPCErrBusinessNonExist         ImportDbServicesPreCheckErr = "所属业务不存在"
	IDBPCErrOptTimeInvalid           ImportDbServicesPreCheckErr = "运维时间不规范"
	IDBPCErrDbTypeInvalid            ImportDbServicesPreCheckErr = "数据源类型不规范或对应插件未安装"
	IDBPCErrOracleServiceNameInvalid ImportDbServicesPreCheckErr = "Oracle服务名错误"
	IDBPCErrDB2DbNameInvalid         ImportDbServicesPreCheckErr = "DB2数据库名错误"
	IDBPCErrRuleTemplateInvalid      ImportDbServicesPreCheckErr = "审核规则模板不存在或数据源类型不匹配"
	IDBPCErrLevelInvalid             ImportDbServicesPreCheckErr = "工作台查询的最高审核等级不规范"
)

type projInfo struct {
	proj          *Project
	ruleTemplates map[string]*RuleTemplate // ruleTemplateName -> RuleTemplate
	business      map[string]struct{}      // businessName ->
}

// GetRuleTemplates request SQLE to get the project's all rule templates
func (d *DBServiceUsecase) GetRuleTemplates(ctx context.Context, projectName string) ([]RuleTemplate, error) {
	target, err := d.dmsProxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		return nil, fmt.Errorf("get target failed: %v", err)
	}

	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reply := &struct {
		v1Base.GenericResp
		Data []RuleTemplate `json:"data"`
	}{}

	// project's rule templates
	url1 := target.URL.String() + "/v1/projects/" + projectName + "/rule_template_tips"
	// default rule templates
	url2 := target.URL.String() + "/v1/rule_template_tips"

	var result []RuleTemplate
	for _, url := range []string{url1, url2} {
		err = pkgHttp.Get(ctx, url, header, nil, reply)
		if err != nil {
			return nil, fmt.Errorf("get rule templates failed: %v", err)
		} else if reply.Code != 0 {
			return nil, fmt.Errorf("get rule templates failed: %v", reply.Message)
		}
		result = append(result, reply.Data...)
	}

	return result, nil
}

func (d *DBServiceUsecase) getActiveProjInfo(ctx context.Context, proj *Project) (*projInfo, error) {
	if proj.Status != ProjectStatusActive {
		return nil, fmt.Errorf("%w project: %s status: %s", IDBPCErrProjNotActive, proj.Name, proj.Status)
	}
	// todo 不建议DMS依赖SQLE功能
	templates, err := d.GetRuleTemplates(ctx, proj.Name)
	if err != nil {
		return nil, err
	}
	d.log.Debugf("project:(%s) GetRuleTemplates: %v", proj.Name, templates)
	info := &projInfo{
		proj:          proj,
		ruleTemplates: make(map[string]*RuleTemplate, len(templates)),
		business:      make(map[string]struct{}, len(proj.Business)),
	}
	for i := range templates {
		info.ruleTemplates[templates[i].RuleTemplateName] = &templates[i]
	}
	for i := range proj.Business {
		info.business[proj.Business[i].Name] = struct{}{}
	}
	return info, nil
}

func (d *DBServiceUsecase) readFromImportCsv(fileContent string) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader([]byte(fileContent)))
	r.FieldsPerRecord = ImportDBServicesCsvColumnsNum
	records, err := r.ReadAll()
	if err != nil || len(records) < 1 {
		return nil, fmt.Errorf("failed to read csv: %v", err)
	}
	return records, nil
}

func (d *DBServiceUsecase) handleImportCsvRecord(ctx context.Context, projectInfoMap map[string]*projInfo, isDmsAdmin bool, line []string) (*BizDBServiceArgs, error) {
	if len(line) != ImportDBServicesCsvColumnsNum {
		return nil, fmt.Errorf("invalid csv record")
	}
	csvDbName := line[0]
	csvProj := line[1]
	csvBus := line[2]
	csvDesc := line[3]
	csvDbType := line[4]
	csvHost := line[5]
	csvPort := line[6]
	csvUser := line[7]
	csvPass := line[8]
	csvOracleService := line[9]
	csvDB2DbName := line[10]
	csvOpsTime := line[11]
	csvRuleTemplate := line[12]
	csvLevel := line[13]

	switch "" {
	case csvDbName, csvProj, csvBus, csvDesc, csvDbType, csvHost, csvPort, csvUser, csvPass, csvRuleTemplate:
		return nil, IDBPCErrInvalidInput
	}

	_, projExist := projectInfoMap[csvProj]
	if isDmsAdmin && !projExist {
		proj, err := d.projectUsecase.GetProjectByName(ctx, csvProj)
		if err != nil {
			d.log.Errorf("get project by name(%s) failed: %v", csvProj, err)
			return nil, fmt.Errorf("%w get project by name(%s) failed: %v", IDBPCErrProjNonExist, csvProj, err)
		}
		info, err := d.getActiveProjInfo(ctx, proj)
		if err != nil {
			return nil, err
		}
		projectInfoMap[proj.Name] = info
	} else if !projExist {
		return nil, fmt.Errorf("%w project name:(%s) is not existent", IDBPCErrProjNotAllowed, csvProj)
	}

	if _, businessExist := projectInfoMap[csvProj].business[csvBus]; !businessExist {
		return nil, fmt.Errorf("%w business name:(%s) proj:(%s)", IDBPCErrBusinessNonExist, csvBus, csvProj)
	}

	var mp pkgPeriods.Periods
	if csvOpsTime != "" {
		var err error
		mp, err = pkgPeriods.ParsePeriods(csvOpsTime)
		if err != nil {
			return nil, fmt.Errorf("%w parse MaintenancePeriod failed: %v", IDBPCErrOptTimeInvalid, err)
		}
	}

	additionalParams, err := d.GetDriverParamsByDBType(ctx, csvDbType)
	if err != nil {
		return nil, fmt.Errorf("%w err:%v", IDBPCErrDbTypeInvalid, err)
	}
	// todo 理论上dms感知不到数据库类型
	if csvDbType == "Oracle" && (csvOracleService == "" || additionalParams == nil) {
		return nil, fmt.Errorf("%w err:%v", IDBPCErrOracleServiceNameInvalid, err)
	} else if csvDbType == "Oracle" && csvOracleService != "" {
		if err = additionalParams.SetParamValue("service_name", csvOracleService); err != nil {
			return nil, fmt.Errorf("%w err:%v", IDBPCErrOracleServiceNameInvalid, err)
		}
	}
	if csvDbType == "DB2" && (csvDB2DbName == "" || additionalParams == nil) {
		return nil, fmt.Errorf("%w err:%v", IDBPCErrDB2DbNameInvalid, err)
	} else if csvDbType == "DB2" && csvDB2DbName != "" {
		if err = additionalParams.SetParamValue("database_name", csvDB2DbName); err != nil {
			return nil, fmt.Errorf("%w err:%v", IDBPCErrDB2DbNameInvalid, err)
		}
	}

	ruleTemplate, ruleTemplateExist := projectInfoMap[csvProj].ruleTemplates[csvRuleTemplate]
	if !ruleTemplateExist || ruleTemplate.DbType != csvDbType {
		return nil, fmt.Errorf("%w rule template name:(%s) project:(%s)", IDBPCErrRuleTemplateInvalid, csvRuleTemplate, csvProj)
	}

	var sqlQueryConfig *SQLQueryConfig
	switch csvLevel {
	case string(v1.AuditLevelNormal), string(v1.AuditLevelNotice), string(v1.AuditLevelWarn), string(v1.AuditLevelError):
		sqlQueryConfig = &SQLQueryConfig{
			MaxPreQueryRows:                  0,
			QueryTimeoutSecond:               0,
			AuditEnabled:                     true,
			AllowQueryWhenLessThanAuditLevel: csvLevel,
		}
	default:
		return nil, fmt.Errorf("%w input:(%s)", IDBPCErrLevelInvalid, csvLevel)
	}

	desc, pass := csvDesc, csvPass
	dbServiceArg := &BizDBServiceArgs{
		Name:              csvDbName,
		Desc:              &desc,
		DBType:            csvDbType,
		Host:              csvHost,
		Port:              csvPort,
		User:              csvUser,
		Password:          &pass,
		Business:          csvBus,
		Source:            string(pkgConst.DBServiceSourceNameSQLE),
		AdditionalParams:  additionalParams,
		ProjectUID:        projectInfoMap[csvProj].proj.UID,
		MaintenancePeriod: mp,
		RuleTemplateName:  ruleTemplate.RuleTemplateName,
		RuleTemplateID:    ruleTemplate.RuleTemplateID,
		SQLQueryConfig:    sqlQueryConfig,
		IsMaskingSwitch:   false, //todo
	}
	return dbServiceArg, nil
}

func (d *DBServiceUsecase) importDBServicesCheck(ctx context.Context, fileContent string, projectInfoMap map[string]*projInfo, isDmsAdmin bool) ([]*BizDBServiceArgs, []byte, error) {
	records, err := d.readFromImportCsv(fileContent)
	if err != nil {
		return nil, nil, err
	}

	buf := &bytes.Buffer{}
	buf.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	csvWriter := csv.NewWriter(buf)
	// 在csv每行末尾增加一列用以描述该行预检问题，最终形成预检审核结果文件
	writeLine := make([]string, 0, ImportDBServicesCsvColumnsNum+1)

	var errLines int
	results := make([]*BizDBServiceArgs, 0, len(records)-1)
	for k, v := range records {
		if k == 0 { // 标题行
			writeLine = append(append(writeLine, records[0]...), "问题")
			if err := csvWriter.Write(writeLine); err != nil {
				return nil, nil, err
			}
			continue
		}

		var lineErr ImportDbServicesPreCheckErr = "无"
		bizDb, err := d.handleImportCsvRecord(ctx, projectInfoMap, isDmsAdmin, v)
		if err != nil && !errors.As(err, &lineErr) {
			return nil, nil, err
		} else if err != nil {
			errLines++
			d.log.Debugf("handleLine %d ImportDbServicesPreCheckErr: %v", k+2, lineErr)
		}

		results = append(results, bizDb)
		writeLine = append(append(writeLine[:0], records[k]...), lineErr.Error())
		if err := csvWriter.Write(writeLine); err != nil {
			return nil, nil, err
		}
	}

	// 预检未通过，生成预检审核结果
	if errLines > 0 {
		csvWriter.Flush()
		return nil, buf.Bytes(), nil
	}

	// 预检通过，返回csv解析结果
	return results, nil, nil
}

func (d *DBServiceUsecase) ImportDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) ([]*BizDBServiceArgs, []byte, error) {
	// 权限校验
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, userUid, projectUid); err != nil {
		return nil, nil, fmt.Errorf("check user is project admin failed: %v", err)
	} else if !isAdmin {
		return nil, nil, fmt.Errorf("user is not project admin")
	}

	proj, err := d.projectUsecase.GetProject(ctx, projectUid)
	if err != nil {
		return nil, nil, err
	}
	info, err := d.getActiveProjInfo(ctx, proj)
	if err != nil {
		return nil, nil, err
	}
	projInfos := make(map[string]*projInfo)
	projInfos[proj.Name] = info

	return d.importDBServicesCheck(ctx, fileContent, projInfos, false)
}

func (d *DBServiceUsecase) ImportDBServicesOfProjectsCheck(ctx context.Context, userUid, fileContent string) ([]*BizDBServiceArgs, []byte, error) {
	// 权限校验
	isAdmin, err := d.opPermissionVerifyUsecase.IsUserDMSAdmin(ctx, userUid)
	if err != nil {
		return nil, nil, fmt.Errorf("check user is dms admin failed: %v", err)
	} else if !isAdmin {
		return nil, nil, fmt.Errorf("user is not dms admin")
	}

	projInfos := make(map[string]*projInfo)

	return d.importDBServicesCheck(ctx, fileContent, projInfos, true)
}

func (d *DBServiceUsecase) importDBServices(ctx context.Context, dbs []*DBService) error {
	for _, v := range dbs {
		dbUid, err := pkgRand.GenStrUid()
		if err != nil {
			return err
		}
		v.UID = dbUid

		// 调用其他服务对数据源进行预检查
		if err = d.pluginUsecase.AddDBServicePreCheck(ctx, v); err != nil {
			return fmt.Errorf("precheck db service failed: %w", err)
		}
	}

	if err := d.repo.SaveDBServices(ctx, dbs); err != nil {
		return err
	}

	for _, ds := range dbs {
		err := d.pluginUsecase.OperateDataResourceHandle(ctx, ds.UID, dmsCommonV1.DataResourceTypeDBService, dmsCommonV1.OperationTypeCreate, dmsCommonV1.OperationTimingAfter)
		if err != nil {
			return fmt.Errorf("plugin handle after craete db_service err: %v", err)
		}
	}
	return nil
}

func (d *DBServiceUsecase) ImportDBServicesOfOneProject(ctx context.Context, dbs []*DBService, uid, projectUid string) error {
	// 权限校验
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, uid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not project admin")
	}

	return d.importDBServices(ctx, dbs)
}

func (d *DBServiceUsecase) ImportDBServicesOfProjects(ctx context.Context, dbs []*DBService, uid string) error { // 权限校验
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserDMSAdmin(ctx, uid); err != nil {
		return fmt.Errorf("check user is dms admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not dms admin")
	}

	return d.importDBServices(ctx, dbs)
}

func (d *DBServiceUsecase) DBServicesConnection(ctx context.Context, dbs []dmsV1.CheckDbsConnectable) (*dmsV1.DBServicesConnectionItem, error) {
	ret := &dmsV1.DBServicesConnectionItem{}
	for _, v := range dbs {
		connectable, err := d.IsConnectable(ctx, v1.CheckDbConnectable{
			DBType:           v.DBType,
			User:             v.User,
			Host:             v.Host,
			Port:             v.Port,
			Password:         v.Password,
			AdditionalParams: v.AdditionalParams,
		})
		if err != nil {
			return nil, err
		}

		ok := true
		for _, c := range connectable {
			if !c.IsConnectable {
				ok = false
				break
			}
		}
		if !ok {
			ret.FailedNum++
			ret.FailedNames = append(ret.FailedNames, v.Name)
		} else {
			ret.SuccessfulNum++
		}
	}
	return ret, nil
}
