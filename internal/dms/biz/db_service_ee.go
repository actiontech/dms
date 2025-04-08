//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"sync"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/locale"
	v1Base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	v1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/dms-common/pkg/config"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	pkgPeriods "github.com/actiontech/dms/pkg/periods"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type RuleTemplate struct {
	RuleTemplateName string `json:"rule_template_name"`
	RuleTemplateID   string `json:"rule_template_id"`
	DbType           string `json:"db_type"`
}

type ImportDbServicesPreCheckErr string

func (e ImportDbServicesPreCheckErr) Error() string {
	return string(e)
}

func localizeIDBPreCheckErr(ctx context.Context, msg *i18n.Message) ImportDbServicesPreCheckErr {
	return ImportDbServicesPreCheckErr(locale.Bundle.LocalizeMsgByCtx(ctx, msg))
}

type projInfo struct {
	proj          *Project
	ruleTemplates map[string]*RuleTemplate // ruleTemplateName -> RuleTemplate
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
		return nil, fmt.Errorf("%w project: %s status: %s", localizeIDBPreCheckErr(ctx, locale.IDBPCErrProjNotActive), proj.Name, proj.Status)
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
	}
	for i := range templates {
		info.ruleTemplates[templates[i].RuleTemplateName] = &templates[i]
	}
	return info, nil
}

type ImportDbServicesCsvRow struct {
	DbName           string `validate:"dbNameFormat"`
	ProjName         string `validate:"required"`
	Business         string `validate:"required"`
	Desc             string `validate:"required"`
	DbType           string `validate:"required"`
	Host             string `validate:"required"`
	Port             string `validate:"required"`
	User             string `validate:"required"`
	Password         string `validate:"required"`
	OracleService    string `validate:"required_if=DbType Oracle"`
	DB2DbName        string `validate:"required_if=DbType DB2"`
	OpsTime          string
	RuleTemplateName string `validate:"required"`
	AuditLevel       string `validate:"oneof='' error warn notice normal"`
}

type ImportDbServicesCheckResultCsvRow struct {
	*ImportDbServicesCsvRow
	Problem string
}

var rowFieldToMsg = map[string]*i18n.Message{
	"DbName":           locale.DBServiceDbName,
	"ProjName":         locale.DBServiceProjName,
	"Business":         locale.DBServiceBusiness,
	"Desc":             locale.DBServiceDesc,
	"DbType":           locale.DBServiceDbType,
	"Host":             locale.DBServiceHost,
	"Port":             locale.DBServicePort,
	"User":             locale.DBServiceUser,
	"Password":         locale.DBServicePassword,
	"OracleService":    locale.DBServiceOracleService,
	"DB2DbName":        locale.DBServiceDB2DbName,
	"OpsTime":          locale.DBServiceOpsTime,
	"RuleTemplateName": locale.DBServiceRuleTemplateName,
	"AuditLevel":       locale.DBServiceAuditLevel,
	"Problem":          locale.DBServiceProblem,
}

func (d *DBServiceUsecase) checkImportCsvRow(ctx context.Context, projectInfoMap map[string]*projInfo, isDmsAdmin bool, row *ImportDbServicesCsvRow) error {
	err := config.Validate(row)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var cols []string
			for _, v := range validationErrors {
				if msg, exist := rowFieldToMsg[v.StructField()]; exist {
					cols = append(cols, locale.Bundle.LocalizeMsgByCtx(ctx, msg))
				} else {
					cols = append(cols, v.StructField()) // rowFieldToMsg 未定义国际化消息时，使用结构的字段名称
				}
			}
			return ImportDbServicesPreCheckErr(fmt.Sprintf(locale.Bundle.LocalizeMsgByCtx(ctx, locale.IDBPCErrMissingOrInvalidCols), strings.Join(cols, ";")))
		}
		return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrInvalidInput), err)
	}

	_, projExist := projectInfoMap[row.ProjName]
	if isDmsAdmin && !projExist {
		proj, err := d.projectUsecase.GetProjectByName(ctx, row.ProjName)
		if err != nil {
			d.log.Errorf("get project by name(%s) failed: %v", row.ProjName, err)
			return fmt.Errorf("%w get project by name(%s) failed: %v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrProjNonExist), row.ProjName, err)
		}
		info, err := d.getActiveProjInfo(ctx, proj)
		if err != nil {
			return err
		}
		projectInfoMap[proj.Name] = info
	} else if !projExist {
		return fmt.Errorf("%w project name:(%s) is not existent", localizeIDBPreCheckErr(ctx, locale.IDBPCErrProjNotAllowed), row.ProjName)
	}

	if row.OpsTime != "" {
		if _, err = pkgPeriods.ParsePeriods(row.OpsTime); err != nil {
			return fmt.Errorf("%w parse MaintenancePeriod failed: %v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrOptTimeInvalid), err)
		}
	}

	additionalParams, err := d.GetDriverParamsByDBType(ctx, row.DbType)
	if err != nil {
		return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrDbTypeInvalid), err)
	}

	if row.DbType == "Oracle" && (row.OracleService == "" || additionalParams == nil) {
		return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrOracleServiceNameInvalid), err)
	} else if row.DbType == "Oracle" && row.OracleService != "" {
		if err = additionalParams.SetParamValue("service_name", row.OracleService); err != nil {
			return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrOracleServiceNameInvalid), err)
		}
	}
	if row.DbType == "DB2" && (row.DB2DbName == "" || additionalParams == nil) {
		return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrDB2DbNameInvalid), err)
	} else if row.DbType == "DB2" && row.DB2DbName != "" {
		if err = additionalParams.SetParamValue("database_name", row.DB2DbName); err != nil {
			return fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrDB2DbNameInvalid), err)
		}
	}

	ruleTemplate, ruleTemplateExist := projectInfoMap[row.ProjName].ruleTemplates[row.RuleTemplateName]
	if !ruleTemplateExist || ruleTemplate.DbType != row.DbType {
		return fmt.Errorf("%w rule template name:(%s) project:(%s)", localizeIDBPreCheckErr(ctx, locale.IDBPCErrRuleTemplateInvalid), row.RuleTemplateName, row.ProjName)
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

func (d *DBServiceUsecase) genImportDbServicesCheckResultCsv(ctx context.Context, inputRows []*ImportDbServicesCsvRow, inputErrs map[int]ImportDbServicesPreCheckErr) ([]byte, error) {
	// 填充预检问题
	resultRows := make([]*ImportDbServicesCheckResultCsvRow, len(inputRows))
	for k := range inputRows {
		problem := locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceNoProblem)
		if _, exist := inputErrs[k]; exist {
			problem = inputErrs[k].Error()
		}
		resultRows[k] = &ImportDbServicesCheckResultCsvRow{
			ImportDbServicesCsvRow: inputRows[k],
			Problem:                problem,
		}
	}

	buff := new(bytes.Buffer)
	buff.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	csvWriter := csv.NewWriter(buff)
	err := csvWriter.Write([]string{
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceDbName),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceProjName),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceBusiness),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceDesc),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceDbType),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceHost),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServicePort),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceUser),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServicePassword),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceOracleService),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceDB2DbName),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceOpsTime),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceRuleTemplateName),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceAuditLevel),
		locale.Bundle.LocalizeMsgByCtx(ctx, locale.DBServiceProblem),
	})
	if err != nil {
		return nil, err
	}
	for _, row := range resultRows {
		err = csvWriter.Write([]string{
			row.DbName,
			row.ProjName,
			row.Business,
			row.Desc,
			row.DbType,
			row.Host,
			row.Port,
			row.User,
			row.Password,
			row.OracleService,
			row.DB2DbName,
			row.OpsTime,
			row.RuleTemplateName,
			row.AuditLevel,
			row.Problem,
		})
		if err != nil {
			return nil, err
		}
	}
	csvWriter.Flush()
	return buff.Bytes(), nil
}

func (d *DBServiceUsecase) genBizDBServiceArgs4Import(ctx context.Context, projectInfoMap map[string]*projInfo, rows []*ImportDbServicesCsvRow) ([]*BizDBServiceArgs, error) {
	bizDBServiceArgs := make([]*BizDBServiceArgs, 0, len(rows))
	for _, row := range rows {
		var mp pkgPeriods.Periods
		if row.OpsTime != "" {
			var err error
			mp, err = pkgPeriods.ParsePeriods(row.OpsTime)
			if err != nil {
				return nil, fmt.Errorf("%w parse MaintenancePeriod failed: %v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrOptTimeInvalid), err)
			}
		}

		additionalParams, err := d.GetDriverParamsByDBType(ctx, row.DbType)
		if err != nil {
			return nil, fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrDbTypeInvalid), err)
		}
		if row.DbType == "Oracle" && additionalParams != nil {
			if err = additionalParams.SetParamValue("service_name", row.OracleService); err != nil {
				return nil, fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrOracleServiceNameInvalid), err)
			}
		}
		if row.DbType == "DB2" && additionalParams != nil {
			if err = additionalParams.SetParamValue("database_name", row.DB2DbName); err != nil {
				return nil, fmt.Errorf("%w err:%v", localizeIDBPreCheckErr(ctx, locale.IDBPCErrDB2DbNameInvalid), err)
			}
		}

		ruleTemplate, ruleTemplateExist := projectInfoMap[row.ProjName].ruleTemplates[row.RuleTemplateName]
		if !ruleTemplateExist || ruleTemplate.DbType != row.DbType {
			return nil, fmt.Errorf("%w rule template name:(%s) project:(%s)", localizeIDBPreCheckErr(ctx, locale.IDBPCErrRuleTemplateInvalid), row.RuleTemplateName, row.ProjName)
		}

		sqlQueryConfig := &SQLQueryConfig{}
		if row.AuditLevel != "" {
			sqlQueryConfig.AuditEnabled = true
			sqlQueryConfig.AllowQueryWhenLessThanAuditLevel = row.AuditLevel
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

func (d *DBServiceUsecase) getImportRowsFromCsvContent(s string) ([]*ImportDbServicesCsvRow, error) {
	reader := csv.NewReader(strings.NewReader(s))
	reader.FieldsPerRecord = 14
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := make([]*ImportDbServicesCsvRow, 0, len(records)-1)
	for k, row := range records {
		if k == 0 {
			continue // 跳过标题
		}
		result = append(result, &ImportDbServicesCsvRow{
			DbName:           row[0],
			ProjName:         row[1],
			Business:         row[2],
			Desc:             row[3],
			DbType:           row[4],
			Host:             row[5],
			Port:             row[6],
			User:             row[7],
			Password:         row[8],
			OracleService:    row[9],
			DB2DbName:        row[10],
			OpsTime:          row[11],
			RuleTemplateName: row[12],
			AuditLevel:       row[13],
		})
	}
	return result, nil
}

func (d *DBServiceUsecase) importDBServicesCheck(ctx context.Context, fileContent string, projectInfoMap map[string]*projInfo, isDmsAdmin bool) ([]*BizDBServiceArgs, []byte, error) {

	inputRows, err := d.getImportRowsFromCsvContent(fileContent)
	if err != nil {
		return nil, nil, err
	}

	// 预检
	checkErrs, err := d.checkImportCsvRows(ctx, projectInfoMap, isDmsAdmin, inputRows)
	if err != nil {
		return nil, nil, err
	}

	// 预检未通过
	if len(checkErrs) > 0 {
		resultCsv, err := d.genImportDbServicesCheckResultCsv(ctx, inputRows, checkErrs)
		return nil, resultCsv, err
	}

	// 预检通过
	bizDBServiceArgs, err := d.genBizDBServiceArgs4Import(ctx, projectInfoMap, inputRows)
	return bizDBServiceArgs, nil, err
}

func (d *DBServiceUsecase) ImportDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) ([]*BizDBServiceArgs, []byte, error) {
	// 权限校验
	if canOpProject, err := d.opPermissionVerifyUsecase.CanOpProject(ctx, userUid, projectUid); err != nil {
		return nil, nil, fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return nil, nil, fmt.Errorf("user is not project admin or golobal op permission user")
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
	hasGlobalOpPermission, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, userUid)
	if err != nil {
		return nil, nil, fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !hasGlobalOpPermission {
		return nil, nil, fmt.Errorf("user is not project admin or golobal op permission user")
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

		err = d.createDBService(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBServiceUsecase) ImportDBServicesOfOneProject(ctx context.Context, dbs []*DBService, uid, projectUid string) error {
	// 权限校验
	if canOpProject, err := d.opPermissionVerifyUsecase.CanOpProject(ctx, uid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	return d.importDBServices(ctx, dbs)
}

func (d *DBServiceUsecase) ImportDBServicesOfProjects(ctx context.Context, dbs []*DBService, uid string) error {
	// 权限校验
	if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, uid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canGlobalOp {
		return fmt.Errorf("user is not project admin or golobal op permission user")
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
	canViewGlobal, err := d.opPermissionVerifyUsecase.CanViewGlobal(ctx, currentUserUid)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check user can view global")
	}

	var userBindProjects []dmsCommonV1.UserBindProject
	if !canViewGlobal {
		projectWithOpPermissions, err := d.opPermissionVerifyUsecase.GetUserProjectOpPermission(ctx, currentUserUid)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get user project with op permission")
		}
		userBindProjects = d.opPermissionVerifyUsecase.GetUserManagerProject(ctx, projectWithOpPermissions)
	} else {
		projects, _, err := d.projectUsecase.ListProject(ctx, &ListProjectsOption{
			PageNumber:   1,
			LimitPerPage: 999,
		}, currentUserUid)
		if err != nil {
			return nil, 0, err
		}
		for _, project := range projects {
			userBindProjects = append(userBindProjects, dmsCommonV1.UserBindProject{ProjectID: project.UID, ProjectName: project.Name, IsManager: true})
		}
	}

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
