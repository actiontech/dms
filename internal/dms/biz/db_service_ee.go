//go:build enterprise

package biz

import (
	"bytes"
	"context"
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
	"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"reflect"
	"strings"
	"sync"
)

var validate = validator.New()
var csvTitleLine string

func init() {
	csvTitle, _ := gocsv.MarshalString([]*ImportDbServicesCsvRow{})
	csvTitleLine = strings.TrimSuffix(csvTitle, "\n")

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("csv")
	})
}

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
	IDBPCErrInvalidInput             ImportDbServicesPreCheckErr = "若无特别说明每列均为必填"
	IDBPCErrProjNonExist             ImportDbServicesPreCheckErr = "所属项目不存在"
	IDBPCErrProjNotActive            ImportDbServicesPreCheckErr = "所属项目状态异常"
	IDBPCErrProjNotAllowed           ImportDbServicesPreCheckErr = "所属项目不是操作中的项目"
	IDBPCErrBusinessNonExist         ImportDbServicesPreCheckErr = "项目业务固定且所属业务不存在"
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

type ImportDbServicesCsvRow struct {
	DbName           string `csv:"数据源名称" validate:"required"`
	ProjName         string `csv:"所属项目(平台已有的项目名称)" validate:"required"`
	Business         string `csv:"所属业务(项目已有的业务名称)" validate:"required"`
	Desc             string `csv:"数据源描述" validate:"required"`
	DbType           string `csv:"数据源类型" validate:"required"`
	Host             string `csv:"数据源地址" validate:"required"`
	Port             string `csv:"数据源端口" validate:"required"`
	User             string `csv:"数据源连接用户" validate:"required"`
	Password         string `csv:"数据源密码" validate:"required"`
	OracleService    string `csv:"服务名(Oracle需填)" validate:"required_if=DbType Oracle"`
	DB2DbName        string `csv:"数据库名(DB2需填)" validate:"required_if=DbType DB2"`
	OpsTime          string `csv:"运维时间(非必填，9:30-11:00;14:10-18:30)"`
	RuleTemplateName string `csv:"审核规则模板(项目已有的规则模板)" validate:"required"`
	AuditLevel       string `csv:"工作台查询的最高审核等级[error|warn|notice|normal]" validate:"oneof=error warn notice normal"`
}

type ImportDbServicesCheckResultCsvRow struct {
	*ImportDbServicesCsvRow
	Problem string `csv:"问题"`
}

func (d *DBServiceUsecase) checkImportCsvRow(ctx context.Context, projectInfoMap map[string]*projInfo, isDmsAdmin bool, row *ImportDbServicesCsvRow) error {
	err := validate.Struct(row)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var cols []string
			for _, v := range validationErrors {
				cols = append(cols, v.Field())
			}
			return ImportDbServicesPreCheckErr(fmt.Sprintf("缺失或不规范的列：%s", strings.Join(cols, ";")))
		}
		return fmt.Errorf("%w err:%v", IDBPCErrInvalidInput, err)
	}

	_, projExist := projectInfoMap[row.ProjName]
	if isDmsAdmin && !projExist {
		proj, err := d.projectUsecase.GetProjectByName(ctx, row.ProjName)
		if err != nil {
			d.log.Errorf("get project by name(%s) failed: %v", row.ProjName, err)
			return fmt.Errorf("%w get project by name(%s) failed: %v", IDBPCErrProjNonExist, row.ProjName, err)
		}
		info, err := d.getActiveProjInfo(ctx, proj)
		if err != nil {
			return err
		}
		projectInfoMap[proj.Name] = info
	} else if !projExist {
		return fmt.Errorf("%w project name:(%s) is not existent", IDBPCErrProjNotAllowed, row.ProjName)
	}

	if _, businessExist := projectInfoMap[row.ProjName].business[row.Business]; !businessExist && projectInfoMap[row.ProjName].proj.IsFixedBusiness {
		return fmt.Errorf("%w business name:(%s) proj:(%s)", IDBPCErrBusinessNonExist, row.Business, row.ProjName)
	}

	if row.OpsTime != "" {
		if _, err = pkgPeriods.ParsePeriods(row.OpsTime); err != nil {
			return fmt.Errorf("%w parse MaintenancePeriod failed: %v", IDBPCErrOptTimeInvalid, err)
		}
	}

	additionalParams, err := d.GetDriverParamsByDBType(ctx, row.DbType)
	if err != nil {
		return fmt.Errorf("%w err:%v", IDBPCErrDbTypeInvalid, err)
	}

	if row.DbType == "Oracle" && (row.OracleService == "" || additionalParams == nil) {
		return fmt.Errorf("%w err:%v", IDBPCErrOracleServiceNameInvalid, err)
	} else if row.DbType == "Oracle" && row.OracleService != "" {
		if err = additionalParams.SetParamValue("service_name", row.OracleService); err != nil {
			return fmt.Errorf("%w err:%v", IDBPCErrOracleServiceNameInvalid, err)
		}
	}
	if row.DbType == "DB2" && (row.DB2DbName == "" || additionalParams == nil) {
		return fmt.Errorf("%w err:%v", IDBPCErrDB2DbNameInvalid, err)
	} else if row.DbType == "DB2" && row.DB2DbName != "" {
		if err = additionalParams.SetParamValue("database_name", row.DB2DbName); err != nil {
			return fmt.Errorf("%w err:%v", IDBPCErrDB2DbNameInvalid, err)
		}
	}

	ruleTemplate, ruleTemplateExist := projectInfoMap[row.ProjName].ruleTemplates[row.RuleTemplateName]
	if !ruleTemplateExist || ruleTemplate.DbType != row.DbType {
		return fmt.Errorf("%w rule template name:(%s) project:(%s)", IDBPCErrRuleTemplateInvalid, row.RuleTemplateName, row.ProjName)
	}

	switch row.AuditLevel {
	case string(v1.AuditLevelNormal), string(v1.AuditLevelNotice), string(v1.AuditLevelWarn), string(v1.AuditLevelError):
	default:
		return fmt.Errorf("%w input:(%s)", IDBPCErrLevelInvalid, row.AuditLevel)
	}

	return nil
}

func (d *DBServiceUsecase) checkImportCsvRows(ctx context.Context, projectInfoMap map[string]*projInfo, isDmsAdmin bool, rows []*ImportDbServicesCsvRow) (map[int]ImportDbServicesPreCheckErr, error) {
	checkErrs := make(map[int]ImportDbServicesPreCheckErr)
	for i, row := range rows {
		var lineErr ImportDbServicesPreCheckErr
		err := d.checkImportCsvRow(ctx, projectInfoMap, isDmsAdmin, row)
		if err != nil && !errors.As(err, &lineErr) {
			return nil, err
		} else if err != nil {
			checkErrs[i] = lineErr
			d.log.Debugf("checkImportCsvRow: %v, ImportDbServicesPreCheckErr: %v", row, lineErr)
		}
	}
	return checkErrs, nil
}

func (d *DBServiceUsecase) genImportDbServicesCheckResultCsv(inputRows []*ImportDbServicesCsvRow, inputErrs map[int]ImportDbServicesPreCheckErr) ([]byte, error) {
	resultRows := make([]*ImportDbServicesCheckResultCsvRow, len(inputRows))
	for k := range inputRows {
		problem := "无"
		if _, exist := inputErrs[k]; exist {
			problem = inputErrs[k].Error()
		}
		resultRows[k] = &ImportDbServicesCheckResultCsvRow{
			ImportDbServicesCsvRow: inputRows[k],
			Problem:                problem,
		}
	}

	data, err := gocsv.MarshalBytes(resultRows)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	buf.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	buf.Write(data)

	return buf.Bytes(), nil
}

func (d *DBServiceUsecase) genBizDBServiceArgs4Import(ctx context.Context, projectInfoMap map[string]*projInfo, rows []*ImportDbServicesCsvRow) ([]*BizDBServiceArgs, error) {
	bizDBServiceArgs := make([]*BizDBServiceArgs, 0, len(rows))
	for _, row := range rows {
		var mp pkgPeriods.Periods
		if row.OpsTime != "" {
			var err error
			mp, err = pkgPeriods.ParsePeriods(row.OpsTime)
			if err != nil {
				return nil, fmt.Errorf("%w parse MaintenancePeriod failed: %v", IDBPCErrOptTimeInvalid, err)
			}
		}

		additionalParams, err := d.GetDriverParamsByDBType(ctx, row.DbType)
		if err != nil {
			return nil, fmt.Errorf("%w err:%v", IDBPCErrDbTypeInvalid, err)
		}
		if row.DbType == "Oracle" && additionalParams != nil {
			if err = additionalParams.SetParamValue("service_name", row.OracleService); err != nil {
				return nil, fmt.Errorf("%w err:%v", IDBPCErrOracleServiceNameInvalid, err)
			}
		}
		if row.DbType == "DB2" && additionalParams != nil {
			if err = additionalParams.SetParamValue("database_name", row.DB2DbName); err != nil {
				return nil, fmt.Errorf("%w err:%v", IDBPCErrDB2DbNameInvalid, err)
			}
		}

		ruleTemplate, ruleTemplateExist := projectInfoMap[row.ProjName].ruleTemplates[row.RuleTemplateName]
		if !ruleTemplateExist || ruleTemplate.DbType != row.DbType {
			return nil, fmt.Errorf("%w rule template name:(%s) project:(%s)", IDBPCErrRuleTemplateInvalid, row.RuleTemplateName, row.ProjName)
		}

		sqlQueryConfig := &SQLQueryConfig{
			MaxPreQueryRows:                  0,
			QueryTimeoutSecond:               0,
			AuditEnabled:                     true,
			AllowQueryWhenLessThanAuditLevel: row.AuditLevel,
		}

		pass, desc := row.Password, row.Desc
		bizDBServiceArgs = append(bizDBServiceArgs, &BizDBServiceArgs{
			Name:              row.DbName,
			Desc:              &desc,
			DBType:            row.DbType,
			Host:              row.Host,
			Port:              row.Port,
			User:              row.User,
			Password:          &pass,
			Business:          row.Business,
			Source:            string(pkgConst.DBServiceSourceNameSQLE),
			AdditionalParams:  additionalParams,
			ProjectUID:        projectInfoMap[row.ProjName].proj.UID,
			MaintenancePeriod: mp,
			RuleTemplateName:  ruleTemplate.RuleTemplateName,
			RuleTemplateID:    ruleTemplate.RuleTemplateID,
			SQLQueryConfig:    sqlQueryConfig,
			IsMaskingSwitch:   false,
		})
	}

	return bizDBServiceArgs, nil
}

func (d *DBServiceUsecase) importDBServicesCheck(ctx context.Context, fileContent string, projectInfoMap map[string]*projInfo, isDmsAdmin bool) ([]*BizDBServiceArgs, []byte, error) {
	if !strings.HasPrefix(strings.TrimPrefix(fileContent, "\xEF\xBB\xBF"), csvTitleLine) {
		return nil, nil, fmt.Errorf("csv title row is invalid, or it's not encoded with UTF-8")
	}

	var inputRows []*ImportDbServicesCsvRow
	if err := gocsv.UnmarshalString(fileContent, &inputRows); err != nil {
		return nil, nil, err
	}

	// 预检
	checkErrs, err := d.checkImportCsvRows(ctx, projectInfoMap, isDmsAdmin, inputRows)
	if err != nil {
		return nil, nil, err
	}

	// 预检未通过
	if len(checkErrs) > 0 {
		resultCsv, err := d.genImportDbServicesCheckResultCsv(inputRows, checkErrs)
		return nil, resultCsv, err
	}

	// 预检通过
	bizDBServiceArgs, err := d.genBizDBServiceArgs4Import(ctx, projectInfoMap, inputRows)
	return bizDBServiceArgs, nil, err
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

func (d *DBServiceUsecase) ImportDBServicesOfProjects(ctx context.Context, dbs []*DBService, uid string) error {
	// 权限校验
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserDMSAdmin(ctx, uid); err != nil {
		return fmt.Errorf("check user is dms admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not dms admin")
	}

	return d.importDBServices(ctx, dbs)
}

func (d *DBServiceUsecase) DBServicesConnection(ctx context.Context, dbs []dmsV1.CheckDbsConnectable) (*dmsV1.DBServicesConnectionItem, error) {
	ret := &dmsV1.DBServicesConnectionItem{}
	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(dbs))
	for k := range dbs {
		go func(v dmsV1.CheckDbsConnectable) {
			defer wg.Done()
			ok := true
			connectable, err := d.IsConnectable(ctx, v1.CheckDbConnectable{
				DBType:           v.DBType,
				User:             v.User,
				Host:             v.Host,
				Port:             v.Port,
				Password:         v.Password,
				AdditionalParams: v.AdditionalParams,
			})
			if err != nil {
				ok = false
				d.log.Errorf("check db connectable err: %v", err)
			} else {
				for _, c := range connectable {
					if !c.IsConnectable {
						ok = false
						break
					}
				}
			}

			mtx.Lock()
			defer mtx.Unlock()
			if !ok {
				ret.FailedNum++
				ret.FailedNames = append(ret.FailedNames, v.Name)
			} else {
				ret.SuccessfulNum++
			}
		}(dbs[k])
	}
	wg.Wait()
	return ret, nil
}

type GlobalDBService struct {
	DBService
	ProjectName           string
	UnfinishedWorkflowNum int64
}

func (d *DBServiceUsecase) ListGlobalDBServices(ctx context.Context, option *ListDBServicesOption, currentUserUid string) (globalDBServices []*GlobalDBService, total int64, err error) {
	projectWithOpPermissions, err := d.opPermissionVerifyUsecase.GetUserProjectOpPermission(ctx, currentUserUid)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user project with op permission")
	}
	userBindProjects := d.opPermissionVerifyUsecase.GetUserManagerProject(ctx, projectWithOpPermissions)

	projectID2Name := make(map[string]string)
	var managedProjectIDs []string
	for k := range userBindProjects {
		if userBindProjects[k].IsManager {
			managedProjectIDs = append(managedProjectIDs, userBindProjects[k].ProjectID)
			projectID2Name[userBindProjects[k].ProjectID] = userBindProjects[k].ProjectName
		}
	}
	if len(managedProjectIDs) == 0 {
		return nil, 0, fmt.Errorf("the user does not manage any project")
	}

	option.FilterBy = append(option.FilterBy, pkgConst.FilterCondition{
		Field:    string(DBServiceFieldProjectUID),
		Operator: pkgConst.FilterOperatorIn,
		Value:    managedProjectIDs,
	})

	dbServices, total, err := d.repo.ListDBServices(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list db services failed: %w", err)
	}

	var unfinishedNumMap map[string]int64
	if len(dbServices) > 0 {
		dbServicesIds := make([]string, 0, len(dbServices))
		for _, v := range dbServices {
			dbServicesIds = append(dbServicesIds, v.UID)
		}
		// todo: 临时方案，直接调用sqle接口获取，后续需调整
		unfinishedNumMap, err = d.getUnfinishedWorkflowsCountOfDBServices(ctx, dbServicesIds)
		if err != nil {
			return nil, 0, err
		}
	}

	globalDBServices = make([]*GlobalDBService, len(dbServices))
	for i, v := range dbServices {
		globalDBServices[i] = &GlobalDBService{
			DBService:             *v,
			ProjectName:           projectID2Name[v.ProjectUID],
			UnfinishedWorkflowNum: unfinishedNumMap[v.UID],
		}
	}

	return globalDBServices, total, nil
}

type GlobalDBServiceTips struct {
	DbType []string
}

func (d *DBServiceUsecase) ListGlobalDBServicesTips(ctx context.Context, currentUserUid string) (*GlobalDBServiceTips, error) {
	var globalDBServiceTips GlobalDBServiceTips
	err := d.repo.GetFieldDistinctValue(ctx, DBServiceFieldDBType, &globalDBServiceTips.DbType)
	return &globalDBServiceTips, err
}

// getUnfinishedWorkflowsCountOfDBServices return map: dbServicesId -> UnfinishedWorkflowNum
func (d *DBServiceUsecase) getUnfinishedWorkflowsCountOfDBServices(ctx context.Context, dbServicesIds []string) (map[string]int64, error) {
	target, err := d.dmsProxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		return nil, fmt.Errorf("get proxy target failed: %v", err)
	}
	url := target.URL.String() + "/v1/workflows/statistic_of_instances?instance_id=" +
		strings.Join(dbServicesIds, "&instance_id=")

	header := map[string]string{"Authorization": pkgHttp.DefaultDMSToken}

	reply := &struct {
		v1Base.GenericResp
		Data []*struct {
			InstanceId      int64 `json:"instance_id"`
			UnfinishedCount int64 `json:"unfinished_count"`
		} `json:"data"`
	}{}

	err = pkgHttp.Get(ctx, url, header, nil, reply)
	if err != nil {
		return nil, fmt.Errorf("get unfinished workflow num err: %v", err)
	} else if reply.Code != 0 {
		return nil, fmt.Errorf("get unfinished workflow num failed: %v", reply.Message)
	}

	result := make(map[string]int64, len(reply.Data))
	for _, v := range reply.Data {
		result[fmt.Sprint(v.InstanceId)] = v.UnfinishedCount
	}

	return result, nil
}
