package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/locale"
	v1Base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	v1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	"github.com/actiontech/dms/pkg/dms-common/pkg/config"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgParams "github.com/actiontech/dms/pkg/params"
	pkgPeriods "github.com/actiontech/dms/pkg/periods"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/go-openapi/strfmt"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"
)

type SQLEConfig struct {
	// DB Service audit enabled
	AuditEnabled bool `json:"audit_enabled"`
	// DB Service rule template name
	RuleTemplateName string `json:"rule_template_name"`
	// DB Service rule template id
	RuleTemplateID string `json:"rule_template_id"`
	// DB Service data export rule template name
	DataExportRuleTemplateName string `json:"data_export_rule_template_name"`
	// DB Service data export rule template id
	DataExportRuleTemplateID string `json:"data_export_rule_template_id"`
	// DB Service SQL query config
	SQLQueryConfig *SQLQueryConfig `json:"sql_query_config"`
}

type LastConnectionStatus string

const (
	LastConnectionStatusSuccess LastConnectionStatus = "connect_success"
	LastConnectionStatusFailed  LastConnectionStatus = "connect_failed"
)

// 数据源
type DBService struct {
	Base

	UID               string                      `json:"uid"`
	Name              string                      `json:"name"`
	Desc              string                      `json:"desc"`
	DBType            string                      `json:"db_type"`
	Host              string                      `json:"host"`
	Port              string                      `json:"port"`
	User              string                      `json:"user"`
	Password          string                      `json:"password"`
	EnvironmentTag    *dmsCommonV1.EnvironmentTag `json:"environment_tag"`
	AdditionalParams  pkgParams.Params            `json:"additional_params"`
	ProjectUID        string                      `json:"project_uid"`
	MaintenancePeriod pkgPeriods.Periods          `json:"maintenance_period"`
	Source            string                      `json:"source"`

	// db service connection
	LastConnectionStatus   *LastConnectionStatus `json:"last_connection_status"`
	LastConnectionTime     *time.Time            `json:"last_connection_time"`
	LastConnectionErrorMsg *string               `json:"last_connection_error_msg"`

	// sqle config
	SQLEConfig      *SQLEConfig `json:"sqle_config"`
	IsMaskingSwitch bool        `json:"is_masking_switch"`
	// PROV config
	AccountPurpose string `json:"account_purpose"`
	// audit plan types
	AuditPlanTypes []*dmsCommonV1.AuditPlanTypes `json:"audit_plan_types"`
	// instance audit plan id
	InstanceAuditPlanID uint   `json:"instance_audit_plan_id"`
	EnableBackup        bool   `json:"enable_backup"`
	BackupMaxRows       uint64 `json:"backup_max_rows"`
}

type DBTypeCount struct {
	DBType string `json:"db_type"`
	Count  int64  `json:"count"`
}

func (d *DBService) GetUID() string {
	return d.UID
}

func (d *DBService) GetRuleTemplateName() string {
	for _, k := range d.AdditionalParams {
		if k.Key == DBServiceAdditionalParam_RuleTemplateName {
			return k.Value
		}
	}
	return ""
}

func (d *DBService) GetAllowQueryWhenLessThanAuditLevel() string {
	if d.SQLEConfig == nil || d.SQLEConfig.SQLQueryConfig == nil {
		return ""
	}
	return d.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel
}

const (
	DBServiceAdditionalParam_RuleTemplateName = "rule_template_name"
)

func newDBService(args *BizDBServiceArgs) (*DBService, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}

	dbService := &DBService{
		UID:               uid,
		Name:              args.Name,
		Desc:              *args.Desc,
		DBType:            args.DBType,
		Host:              args.Host,
		Port:              args.Port,
		User:              args.User,
		Password:          *args.Password,
		AdditionalParams:  args.AdditionalParams,
		ProjectUID:        args.ProjectUID,
		Source:            args.Source,
		MaintenancePeriod: args.MaintenancePeriod,
		SQLEConfig:        &SQLEConfig{},
		IsMaskingSwitch:   args.IsMaskingSwitch,
		EnableBackup:      args.EnableBackup,
		BackupMaxRows:     args.BackupMaxRows,
	}

	if args.AuditEnabled {
		dbService.SQLEConfig.AuditEnabled = args.AuditEnabled
		dbService.SQLEConfig.RuleTemplateID = args.RuleTemplateID
		dbService.SQLEConfig.RuleTemplateName = args.RuleTemplateName
		dbService.SQLEConfig.DataExportRuleTemplateName = args.DataExportRuleTemplateName
		dbService.SQLEConfig.DataExportRuleTemplateID = args.DataExportRuleTemplateID
	}
	if args.EnvironmentTagUID != "" {
		dbService.EnvironmentTag = &dmsCommonV1.EnvironmentTag{
			UID: args.EnvironmentTagUID,
		}
	}

	if args.SQLQueryConfig != nil {
		dbService.SQLEConfig.SQLQueryConfig = args.SQLQueryConfig
	}

	return dbService, nil
}

type AdditionalParams pkgParams.Param

type DBServiceRepo interface {
	SaveDBServices(ctx context.Context, dbService []*DBService) error
	GetDBServicesByIds(ctx context.Context, dbServiceIds []string) (services []*DBService, err error)
	ListDBServices(ctx context.Context, opt *ListDBServicesOption) (services []*DBService, total int64, err error)
	DelDBService(ctx context.Context, dbServiceUid string) error
	GetDBService(ctx context.Context, dbServiceUid string) (*DBService, error)
	GetDBServices(ctx context.Context, conditions []pkgConst.FilterCondition) (services []*DBService, err error)
	CheckDBServiceExist(ctx context.Context, dbServiceUids []string) (exists bool, err error)
	UpdateDBService(ctx context.Context, dbService *DBService) error
	CountDBService(ctx context.Context) ([]DBTypeCount, error)
	GetBusinessByProjectUID(ctx context.Context, projectUid string) ([]string, error)
	GetFieldDistinctValue(ctx context.Context, field DBServiceField, results interface{}) error
}

type DBServiceUsecase struct {
	repo                      DBServiceRepo
	dmsProxyTargetRepo        ProxyTargetRepo
	pluginUsecase             *PluginUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	projectUsecase            *ProjectUsecase
	environmentTagUsecase     *EnvironmentTagUsecase
	log                       *utilLog.Helper
}

func NewDBServiceUsecase(log utilLog.Logger, repo DBServiceRepo, pluginUsecase *PluginUsecase, opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
	projectUsecase *ProjectUsecase, proxyTargetRepo ProxyTargetRepo, environmentTagUsecase *EnvironmentTagUsecase) *DBServiceUsecase {
	return &DBServiceUsecase{
		repo:                      repo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		pluginUsecase:             pluginUsecase,
		projectUsecase:            projectUsecase,
		dmsProxyTargetRepo:        proxyTargetRepo,
		environmentTagUsecase:     environmentTagUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.dbService")),
	}
}

type BizDBServiceArgs struct {
	Name               string
	Desc               *string
	DBType             string
	Host               string
	Port               string
	User               string
	Password           *string
	EnvironmentTagUID  string
	EnvironmentTagName string
	Source             string
	AdditionalParams   pkgParams.Params
	ProjectUID         string
	MaintenancePeriod  pkgPeriods.Periods
	// sqle config
	AuditEnabled               bool
	RuleTemplateName           string
	RuleTemplateID             string
	DataExportRuleTemplateName string
	DataExportRuleTemplateID   string
	SQLQueryConfig             *SQLQueryConfig
	IsMaskingSwitch            bool
	EnableBackup               bool
	BackupMaxRows              uint64
}

type SQLQueryConfig struct {
	MaxPreQueryRows                  int    `json:"max_pre_query_rows"`
	QueryTimeoutSecond               int    `json:"query_timeout_second"`
	AuditEnabled                     bool   `json:"audit_enabled"`
	WorkflowExecEnabled              bool   `json:"workflow_exec_enabled"`
	AllowQueryWhenLessThanAuditLevel string `json:"allow_query_when_less_than_audit_level"`
	RuleTemplateID                   string `json:"rule_template_id"`
	RuleTemplateName                 string `json:"rule_template_name"`
}

func (d *DBServiceUsecase) CreateDBService(ctx context.Context, args *BizDBServiceArgs, currentUserUid string) (uid string, err error) {
	// 检查项目是否归档/删除
	if err := d.projectUsecase.isProjectActive(ctx, args.ProjectUID); err != nil {
		return "", fmt.Errorf("create db service error: %v", err)
	}
	// 检查当前用户有项目管理员权限或者数据源管理权限
	hasPermission, err := d.opPermissionVerifyUsecase.HasManagePermission(ctx, currentUserUid, args.ProjectUID, pkgConst.UIdOfOpPermissionManageProjectDataSource)
	if err != nil {
		return "", fmt.Errorf("check user has permission manageDataSource: %v", err)
	}

	if !hasPermission {
		return "", fmt.Errorf("user is not admin or data source management permission")
	}

	ds, err := newDBService(args)
	if err != nil {
		return "", fmt.Errorf("new db service failed: %w", err)
	}

	err = d.createDBService(ctx, ds)
	if err != nil {
		return "", err
	}
	return ds.UID, nil
}

func (d *DBServiceUsecase) createDBService(ctx context.Context, dbService *DBService) error {
	// 数据源名称格式校验，命名不规范会造成的问题：https://github.com/actiontech/sqle/issues/2810
	if !config.DbNameFormatPattern.MatchString(dbService.Name) {
		return fmt.Errorf("db service name %s is not valid", dbService.Name)
	}

	// 调用其他服务对数据源进行预检查
	if err := d.pluginUsecase.AddDBServicePreCheck(ctx, dbService); err != nil {
		return fmt.Errorf("precheck db service failed: %w", err)
	}

	if err := d.repo.SaveDBServices(ctx, []*DBService{dbService}); err != nil {
		return err
	}

	err := d.pluginUsecase.AddDBServiceAfterHandle(ctx, dbService.UID)
	if err != nil {
		return fmt.Errorf("plugin handle after craete db_service err: %v", err)
	}
	return nil
}

type ListDBServicesOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      DBServiceField
	FilterBy     []pkgConst.FilterCondition
}

func (d *DBServiceUsecase) ListDBService(ctx context.Context, option *ListDBServicesOption, projectUid, currentUserUid string) (dbServices []*DBService, total int64, err error) {
	services, total, err := d.repo.ListDBServices(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list db services failed: %w", err)
	}
	// 只允许系统用户和平台管理/查看权限用户查询所有数据源,同步数据到其他服务(provision)
	if projectUid == "" {
		canViewProject, err := d.opPermissionVerifyUsecase.CanViewProject(ctx, currentUserUid, projectUid, pkgConst.UIdOfOpPermissionManageProjectDataSource)
		if err != nil {
			return nil, 0, err
		}

		if !(currentUserUid == pkgConst.UIDOfUserSys || canViewProject) {
			return nil, 0, fmt.Errorf("user is not sys user or global management or view permission user")
		}
	} else {
		err = d.AddInstanceAuditPlanForDBServiceFromSqle(ctx, projectUid, services)
		if err != nil {
			d.log.Warn("get instance audit Plan from sqle: %w", err)
		}
	}

	return services, total, nil
}

func (d *DBServiceUsecase) TestDbServiceConnections(ctx context.Context, DBServiceList []*DBService, currentUserUid string) []dmsV1.DBServiceIsConnectableReply {
	connectionResp := make([]dmsV1.DBServiceIsConnectableReply, len(DBServiceList))
	concurrencyLimit := make(chan int, 3)
	var wg sync.WaitGroup
	var respMu sync.Mutex

	for i, dbService := range DBServiceList {
		wg.Add(1)

		go func(dbService *DBService, index int) {
			defer func() {
				if r := recover(); r != nil {
					d.log.Errorf("CheckDBServiceIsConnectableByIds panic: %v", r)
				}
			}()
			defer wg.Done()

			if dbService == nil {
				return
			}

			concurrencyLimit <- 1

			connectionResult, err := d.TestDbServiceConnection(ctx, dbService)
			if err != nil {
				d.log.Errorf("db connectionResult uid: %v,TestDBServiceConnection err: %v", connectionResult.DBServiceUid, err)
			}

			dbService.LastConnectionStatus = &connectionResult.ConnectionStatus
			dbService.LastConnectionTime = &connectionResult.TestConnectionTime
			dbService.LastConnectionErrorMsg = &connectionResult.ConnectErrorMessage

			err = d.UpdateDBService(ctx, dbService, currentUserUid)
			if err != nil {
				d.log.Errorf("dbService name: %v,UpdateDBServiceByBiz err: %v", dbService.Name, err)
			}

			respMu.Lock()
			connectionResp[index] = dmsV1.DBServiceIsConnectableReply{
				DBServiceUid:        connectionResult.DBServiceUid,
				ConnectionStatus:    dmsCommonV1.LastConnectionTestStatus(connectionResult.ConnectionStatus),
				TestConnectionTime:  strfmt.DateTime(connectionResult.TestConnectionTime),
				ConnectErrorMessage: connectionResult.ConnectErrorMessage,
			}
			respMu.Unlock()

			<-concurrencyLimit
		}(dbService, i)
	}

	wg.Wait()

	return connectionResp
}

type instanceAuditPlanReply struct {
	Code    int                 `json:"code" example:"0"`
	Message string              `json:"message" example:"ok"`
	Data    []InstanceAuditPlan `json:"data"`
}

type InstanceAuditPlan struct {
	InstanceAuditPlanId uint                          `json:"instance_audit_plan_id"`
	InstanceName        string                        `json:"instance_name"`
	Business            string                        `json:"business"`
	InstanceType        string                        `json:"instance_type"`
	AuditPlanTypes      []*dmsCommonV1.AuditPlanTypes `json:"audit_plan_types"`
}

// TODO 临时实现, 当前请求获取扫描任务的url和参数写死
func (d *DBServiceUsecase) AddInstanceAuditPlanForDBServiceFromSqle(ctx context.Context, projectUid string, dbServices []*DBService) error {
	project, err := d.projectUsecase.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project failed: %v", err)
	}
	target, err := d.dmsProxyTargetRepo.GetProxyTargetByName(ctx, _const.SqleComponentName)
	if err != nil {
		return fmt.Errorf("get proxy target by name failed: %v", err)
	}
	sqleAddr := fmt.Sprintf("%s/v2/projects/%s/instance_audit_plans", target.URL.String(), project.Name)
	header := map[string]string{
		"Authorization":           pkgHttp.DefaultDMSToken,
		i18nPkg.AcceptLanguageKey: locale.Bundle.GetLangTagFromCtx(ctx).String(),
	}
	reqBody := struct {
		PageIndex uint32 `json:"page_index"`
		PageSize  uint32 `json:"page_size"`
	}{
		PageIndex: 1,
		PageSize:  999,
	}
	reply := &instanceAuditPlanReply{}
	if err = pkgHttp.Get(ctx, sqleAddr, header, reqBody, reply); err != nil {
		return fmt.Errorf("get instance audit plan from sqle failed: %v", err)
	}
	if reply.Code != 0 {
		return fmt.Errorf("get instance audit plan from sqle reply code(%v) error: %v", reply.Code, reply.Message)
	}

	for _, dbService := range dbServices {
		for _, instAuditPlan := range reply.Data {
			if dbService.Name == instAuditPlan.InstanceName && dbService.DBType == instAuditPlan.InstanceType {
				dbService.InstanceAuditPlanID = instAuditPlan.InstanceAuditPlanId
				dbService.AuditPlanTypes = instAuditPlan.AuditPlanTypes
			}
		}
	}
	return nil
}

func (d *DBServiceUsecase) ListDBServiceTips(ctx context.Context, req *dmsV1.ListDBServiceTipsReq, userId string) ([]*DBService, error) {
	conditions := []pkgConst.FilterCondition{
		{
			Field:    string(DBServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		},
	}

	if req.FilterDBType != "" {
		conditions = append(conditions, pkgConst.FilterCondition{
			Field:    string(DBServiceFieldDBType),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterDBType,
		})
	}

	dbServices, err := d.repo.GetDBServices(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("list db service tips failed: %w", err)
	}

	if req.FunctionalModule == "" {
		return dbServices, nil
	}
	dbServices = filterExportSupportedDb(dbServices, dmsCommonV1.OpPermissionType(req.FunctionalModule))
	isAdmin, err := d.opPermissionVerifyUsecase.CanViewProject(ctx, userId, req.ProjectUid, "")
	if err != nil {
		return nil, fmt.Errorf("check user is project admin or golobal view permission failed: %v", err)
	}

	if isAdmin {
		return dbServices, nil
	}

	permissions, err := d.opPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, userId, req.ProjectUid)
	if err != nil {
		return nil, err
	}

	ret := make([]*DBService, 0)
	for _, item := range dbServices {
		permissionId, err := pkgConst.ConvertPermissionTypeToId(dmsCommonV1.OpPermissionType(req.FunctionalModule))
		if err != nil {
			return nil, err
		}

		if d.opPermissionVerifyUsecase.UserCanOpDB(permissions, []string{permissionId}, item.UID) {
			ret = append(ret, item)
		}
	}

	return ret, nil
}

func filterExportSupportedDb(dbServices []*DBService, opPermissionType dmsCommonV1.OpPermissionType) []*DBService {
	if opPermissionType != dmsCommonV1.OpPermissionTypeExportCreate {
		return dbServices
	}
	ret := make([]*DBService, 0)
	for _, item := range dbServices {
		_, ok := pkgConst.SupportedDataExportDBTypes[pkgConst.DBType(item.DBType)]
		if ok {
			ret = append(ret, item)
		}
	}
	return ret
}

func (d *DBServiceUsecase) ListDBServiceDriverOption(ctx context.Context) ([]*dmsV1.DatabaseDriverOption, error) {
	options, err := d.pluginUsecase.GetDatabaseDriverOptionsCache(ctx)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func (d *DBServiceUsecase) GetDriverParamsByDBType(ctx context.Context, dbType string) (pkgParams.Params, error) {
	databaseOptions, err := d.ListDBServiceDriverOption(ctx)
	if err != nil {
		return nil, err
	}
	for _, driverOptions := range databaseOptions {
		if driverOptions.DBType == dbType {
			return convertAdditionParamsToParams(driverOptions.Params), nil
		}

	}
	return nil, fmt.Errorf("db type %v is not support", dbType)
}

func convertAdditionParamsToParams(additionalParam []*dmsV1.DatabaseDriverAdditionalParam) pkgParams.Params {
	params := make(pkgParams.Params, len(additionalParam))
	for i, item := range additionalParam {
		params[i] = &pkgParams.Param{
			Key:   item.Name,
			Value: item.Value,
			Desc:  item.Description,
			Type:  pkgParams.ParamType(item.Type),
		}
	}
	return params
}

func (d *DBServiceUsecase) GetActiveDBServices(ctx context.Context, dbServiceIds []string) (dbServices []*DBService, err error) {
	services, err := d.repo.GetDBServicesByIds(ctx, dbServiceIds)
	if err != nil {
		return nil, fmt.Errorf("list db services failed: %w", err)
	}

	for _, service := range services {
		if err = d.projectUsecase.isProjectActive(ctx, service.ProjectUID); err == nil {
			dbServices = append(dbServices, service)
		}
	}

	return dbServices, nil
}

func (d *DBServiceUsecase) GetDBServiceFingerprint(dbService *DBService) string {
	return fmt.Sprintf(`
{
    "id":"%s",
    "host":"%s",
    "port":"%s",
    "user":"%s",
    "password":"%s",
    "params":"%v"
}
`, dbService.UID, dbService.Host, dbService.Port, dbService.User, aes.Md5(dbService.Password), dbService.AdditionalParams)
}

func (d *DBServiceUsecase) DelDBService(ctx context.Context, dbServiceUid, currentUserUid string) (err error) {
	ds, err := d.repo.GetDBService(ctx, dbServiceUid)
	if err != nil {
		return fmt.Errorf("get db service failed: %v", err)
	}
	// 检查项目是否归档/删除
	if err := d.projectUsecase.isProjectActive(ctx, ds.ProjectUID); err != nil {
		return fmt.Errorf("delete db service error: %v", err)
	}
	if canOpProject, err := d.opPermissionVerifyUsecase.HasManagePermission(ctx, currentUserUid, ds.ProjectUID, pkgConst.UIdOfOpPermissionManageProjectDataSource); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	err = d.pluginUsecase.DelDBServicePreCheck(ctx, ds.UID)
	if err != nil {
		return fmt.Errorf("plugin handle before delete db_service err: %v", err)
	}

	if err := d.repo.DelDBService(ctx, dbServiceUid); nil != err {
		return fmt.Errorf("delete data service error: %v", err)
	}

	err = d.pluginUsecase.DelDBServiceAfterHandle(ctx, ds.UID)
	if err != nil {
		return fmt.Errorf("plugin handle after delete db_service err: %v", err)
	}

	return nil
}

func (d *DBServiceUsecase) CheckDBServiceExist(ctx context.Context, dbServiceUids []string) (bool, error) {
	return d.repo.CheckDBServiceExist(ctx, dbServiceUids)
}

func (d *DBServiceUsecase) GetDBService(ctx context.Context, dbServiceUid string) (*DBService, error) {
	return d.repo.GetDBService(ctx, dbServiceUid)
}

type TestDbServiceConnectionResult struct {
	DBServiceUid        string
	ConnectionStatus    LastConnectionStatus
	TestConnectionTime  time.Time
	ConnectErrorMessage string
}

func (d *DBServiceUsecase) TestDbServiceConnection(ctx context.Context, dbService *DBService) (TestDbServiceConnectionResult, error) {
	connectionResult := TestDbServiceConnectionResult{
		DBServiceUid:       dbService.UID,
		TestConnectionTime: time.Now(),
		ConnectionStatus:   LastConnectionStatusSuccess,
	}

	var additionParams []*dmsCommonV1.AdditionalParam
	for _, item := range dbService.AdditionalParams {
		additionParams = append(additionParams, &dmsCommonV1.AdditionalParam{
			Name:  item.Key,
			Value: item.Value,
		})
	}

	checkDbConnectableParams := dmsCommonV1.CheckDbConnectable{
		DBType:           dbService.DBType,
		User:             dbService.User,
		Host:             dbService.Host,
		Port:             dbService.Port,
		Password:         dbService.Password,
		AdditionalParams: additionParams,
	}

	connectable, err := d.IsConnectable(ctx, checkDbConnectableParams)
	if err != nil {
		connectionResult.ConnectionStatus = LastConnectionStatusFailed
		connectionResult.ConnectErrorMessage = err.Error()
		return connectionResult, err
	}

	if len(connectable) == 0 {
		connectionResult.ConnectionStatus = LastConnectionStatusFailed
		connectionResult.ConnectErrorMessage = "check db connectable failed"
	} else {
		for _, c := range connectable {
			if !c.IsConnectable {
				connectionResult.ConnectionStatus = LastConnectionStatusFailed
				connectionResult.ConnectErrorMessage = c.ConnectErrorMessage
				break
			}
		}
	}

	return connectionResult, nil
}

func (d *DBServiceUsecase) UpdateDBService(ctx context.Context, ds *DBService, currentUserUid string) (err error) {
	// 检查项目是否归档/删除
	if err := d.projectUsecase.isProjectActive(ctx, ds.ProjectUID); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}

	if canOpProject, err := d.opPermissionVerifyUsecase.HasManagePermission(ctx, currentUserUid, ds.ProjectUID, pkgConst.UIdOfOpPermissionManageProjectDataSource); err != nil {
		return fmt.Errorf("check user has update data source op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user has no update data source permission")
	}

	if err := d.repo.UpdateDBService(ctx, ds); nil != err {
		return fmt.Errorf("update db service error: %v", err)
	}

	err = d.pluginUsecase.UpdateDBServiceAfterHandle(ctx, ds.UID)
	if err != nil {
		return fmt.Errorf("plugin handle after update db_service err: %v", err)
	}

	return nil
}

func (d *DBServiceUsecase) UpdateDBServiceByArgs(ctx context.Context, dbServiceUid string, updateDBService *BizDBServiceArgs, currentUserUid string) (err error) {
	ds, err := d.repo.GetDBService(ctx, dbServiceUid)
	if err != nil {
		return fmt.Errorf("get db service failed: %v", err)
	}
	// 检查项目是否归档/删除
	if err := d.projectUsecase.isProjectActive(ctx, ds.ProjectUID); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}
	// 检查当前用户有项目管理员权限
	if canOpProject, err := d.opPermissionVerifyUsecase.HasManagePermission(ctx, currentUserUid, ds.ProjectUID, pkgConst.UIdOfOpPermissionManageProjectDataSource); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	// check
	{
		if ds.DBType != updateDBService.DBType {
			return fmt.Errorf("update db service db type is unsupported")
		}

		if updateDBService.Host == "" || updateDBService.Port == "" ||
			updateDBService.User == "" || updateDBService.EnvironmentTagUID == "" {
			return fmt.Errorf("db service's host,port,user,environment can't be empty")
		}
		_, err := d.environmentTagUsecase.GetEnvironmentTagByUID(ctx, updateDBService.EnvironmentTagUID)
		if err != nil {
			return fmt.Errorf("check get environment tag by uid failed: %v", err)
		}
	}
	// update
	{
		if updateDBService.Desc != nil {
			ds.Desc = *updateDBService.Desc
		}
		if updateDBService.Password != nil {
			if *updateDBService.Password == "" {
				return fmt.Errorf("password can't be empty")
			}
			ds.Password = *updateDBService.Password
		}

		ds.Host = updateDBService.Host
		ds.Port = updateDBService.Port
		ds.User = updateDBService.User
		ds.AdditionalParams = updateDBService.AdditionalParams
		ds.MaintenancePeriod = updateDBService.MaintenancePeriod
		ds.IsMaskingSwitch = updateDBService.IsMaskingSwitch
		ds.EnableBackup = updateDBService.EnableBackup
		ds.BackupMaxRows = updateDBService.BackupMaxRows
		ds.SQLEConfig = &SQLEConfig{}
		// 支持新增和更新sqleConfig，不允许删除sqle配置
		if updateDBService.RuleTemplateName != "" {
			ds.SQLEConfig.AuditEnabled = updateDBService.AuditEnabled
			ds.SQLEConfig.RuleTemplateID = updateDBService.RuleTemplateID
			ds.SQLEConfig.RuleTemplateName = updateDBService.RuleTemplateName
			ds.SQLEConfig.DataExportRuleTemplateName = updateDBService.DataExportRuleTemplateName
			ds.SQLEConfig.DataExportRuleTemplateID = updateDBService.DataExportRuleTemplateID
		}
		ds.EnvironmentTag = &dmsCommonV1.EnvironmentTag{
			UID: updateDBService.EnvironmentTagUID,
		}

		if updateDBService.SQLQueryConfig != nil {
			ds.SQLEConfig.SQLQueryConfig = updateDBService.SQLQueryConfig
		}
	}

	if err := d.repo.UpdateDBService(ctx, ds); nil != err {
		return fmt.Errorf("update db service error: %v", err)
	}

	err = d.pluginUsecase.UpdateDBServiceAfterHandle(ctx, ds.UID)
	if err != nil {
		return fmt.Errorf("plugin handle after update db_service err: %v", err)
	}

	return nil
}

type IsConnectableReply struct {
	IsConnectable       bool   `json:"is_connectable"`
	Component           string `json:"component"`
	ConnectErrorMessage string `json:"connect_error_message"`
}

func (d *DBServiceUsecase) IsConnectable(ctx context.Context, params dmsCommonV1.CheckDbConnectable) ([]*IsConnectableReply, error) {
	dmsProxyTargets, err := d.dmsProxyTargetRepo.ListProxyTargetsByScenarios(ctx, []ProxyScenario{ProxyScenarioInternalService})
	if err != nil {
		return nil, err
	}

	ret := make([]*IsConnectableReply, len(dmsProxyTargets))

	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	uri := v1.GetDBConnectionAbleRouter()

	var wg = &sync.WaitGroup{}
	wg.Add(len(dmsProxyTargets))

	for i, target := range dmsProxyTargets {
		go func(i int, target *ProxyTarget) {
			defer wg.Done()

			isConnectableReply := &IsConnectableReply{Component: target.Name}
			var reply = &v1Base.GenericResp{}
			err = pkgHttp.POST(ctx, fmt.Sprintf("%s%s", target.URL.String(), uri), header, params, reply)
			if err != nil {
				isConnectableReply.ConnectErrorMessage = err.Error()
			} else if reply.Code != 0 {
				isConnectableReply.ConnectErrorMessage = reply.Message
			} else {
				isConnectableReply.IsConnectable = true
			}

			ret[i] = isConnectableReply
		}(i, target)
	}

	wg.Wait()

	return ret, nil
}

func (d *DBServiceUsecase) CountDBService(ctx context.Context) ([]DBTypeCount, error) {
	counts, err := d.repo.CountDBService(ctx)
	if err != nil {
		return nil, fmt.Errorf("count db services failed: %w", err)
	}
	return counts, nil
}

func (d *DBServiceUsecase) GetBizDBWithNameByUids(ctx context.Context, uids []string) []UIdWithName {
	if len(uids) == 0 {
		return []UIdWithName{}
	}
	uidWithNameCacheCache.ulock.Lock()
	defer uidWithNameCacheCache.ulock.Unlock()
	if uidWithNameCacheCache.DBCache == nil {
		uidWithNameCacheCache.DBCache = make(map[string]UIdWithName)
	}
	ret := make([]UIdWithName, 0)
	for _, uid := range uids {
		dbCache, ok := uidWithNameCacheCache.DBCache[uid]
		if !ok {
			dbCache = UIdWithName{
				Uid: uid,
			}
			db, err := d.repo.GetDBService(ctx, uid)
			if err == nil {
				dbCache.Name = db.Name
				uidWithNameCacheCache.DBCache[db.UID] = dbCache
			} else {
				d.log.Errorf("get db service for cache err: %v", err)
			}
		}
		ret = append(ret, dbCache)
	}
	return ret
}

func (d *DBServiceUsecase) GetBusiness(ctx context.Context, projectUid string) ([]string, error) {
	business, err := d.repo.GetBusinessByProjectUID(ctx, projectUid)
	if err != nil {
		return nil, fmt.Errorf("get business failed: %v", err)
	}

	return business, nil
}

type CheckDBServicePrivileges struct {
	ComponentPrivilegesResult []*IsConnectableReply
}

func (d *DBServiceUsecase) CheckDBServiceHasEnoughPrivileges(ctx context.Context, params []dmsCommonV1.CheckDbConnectable) ([]*CheckDBServicePrivileges, error) {
	type resultItem struct {
		index  int
		result *CheckDBServicePrivileges
		err    error
	}

	ret := make([]*CheckDBServicePrivileges, len(params))
	resultCh := make(chan resultItem, len(params))
	var wg sync.WaitGroup

	maxConcurrency := 8
	semaphore := make(chan struct{}, maxConcurrency)

	for i, v := range params {
		wg.Add(1)
		go func(i int, param dmsCommonV1.CheckDbConnectable) {
			defer wg.Done()

			semaphore <- struct{}{}        // acquire slot
			defer func() { <-semaphore }() // release slot

			r, err := d.IsConnectable(ctx, param)
			if err != nil {
				resultCh <- resultItem{index: i, err: fmt.Errorf("check db service privileges failed: %v", err)}
				return
			}

			resultCh <- resultItem{
				index: i,
				result: &CheckDBServicePrivileges{
					ComponentPrivilegesResult: r,
				},
			}
		}(i, v)
	}

	wg.Wait()
	close(resultCh)

	for item := range resultCh {
		if item.err != nil {
			return nil, item.err
		}
		ret[item.index] = item.result
	}

	return ret, nil
}
