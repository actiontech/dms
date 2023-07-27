package biz

import (
	"context"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/pkg/database"
	pkgParams "github.com/actiontech/dms/pkg/params"
	pkgPeriods "github.com/actiontech/dms/pkg/periods"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"
)

type SQLEConfig struct {
	// DB Service rule template name
	RuleTemplateName string `json:"rule_template_name"`
	// DB Service rule template id
	RuleTemplateID string `json:"rule_template_id"`
	// DB Service SQL query config
	SQLQueryConfig *SQLQueryConfig `json:"sql_query_config"`
}

// 数据源
type DBService struct {
	Base

	UID                    string
	Name                   string
	Desc                   string
	DBType                 pkgConst.DBType
	Host                   string
	Port                   string
	AdminUser              string
	AdminPassword          string
	EncryptedAdminPassword string
	Business               string
	AdditionalParams       pkgParams.Params
	NamespaceUID           string
	MaintenancePeriod      pkgPeriods.Periods
	Source                 string

	// sqle config
	SQLEConfig *SQLEConfig
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
		AdminUser:         args.AdminUser,
		AdminPassword:     *args.AdminPassword,
		AdditionalParams:  args.AdditionalParams,
		NamespaceUID:      args.NamespaceUID,
		Business:          args.Business,
		MaintenancePeriod: args.MaintenancePeriod,
	}
	if args.SQLQueryConfig != nil && args.RuleTemplateName != "" {
		dbService.SQLEConfig = &SQLEConfig{
			SQLQueryConfig:   args.SQLQueryConfig,
			RuleTemplateName: args.RuleTemplateName,
			RuleTemplateID:   args.RuleTemplateID,
		}
	}
	return dbService, nil
}

type AdditionalParams pkgParams.Param

type DBServiceRepo interface {
	SaveDBService(ctx context.Context, dbService *DBService) error
	GetDBServicesByIds(ctx context.Context, dbServiceIds []string) (services []*DBService, err error)
	ListDBServices(ctx context.Context, opt *ListDBServicesOption) (services []*DBService, total int64, err error)
	DelDBService(ctx context.Context, dbServiceUid string) error
	GetDBService(ctx context.Context, dbServiceUid string) (*DBService, error)
	CheckDBServiceExist(ctx context.Context, dbServiceUids []string) (exists bool, err error)
	UpdateDBService(ctx context.Context, dbService *DBService) error
}

type DBServiceUsecase struct {
	repo                      DBServiceRepo
	pluginUsecase             *PluginUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	namespaceUsecase          *NamespaceUsecase
}

func NewDBServiceUsecase(repo DBServiceRepo, pluginUsecase *PluginUsecase, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, namespaceUsecase *NamespaceUsecase) *DBServiceUsecase {
	return &DBServiceUsecase{
		repo:                      repo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		pluginUsecase:             pluginUsecase,
		namespaceUsecase:          namespaceUsecase,
	}
}

type BizDBServiceArgs struct {
	Name              string
	Desc              *string
	DBType            pkgConst.DBType
	Host              string
	Port              string
	AdminUser         string
	AdminPassword     *string
	Business          string
	AdditionalParams  pkgParams.Params
	NamespaceUID      string
	MaintenancePeriod pkgPeriods.Periods
	// sqle config
	RuleTemplateName string
	RuleTemplateID   string
	SQLQueryConfig   *SQLQueryConfig
}

type SQLQueryConfig struct {
	MaxPreQueryRows                  int    `json:"max_pre_query_rows"`
	QueryTimeoutSecond               int    `json:"query_timeout_second"`
	AuditEnabled                     bool   `json:"audit_enabled"`
	AllowQueryWhenLessThanAuditLevel string `json:"allow_query_when_less_than_audit_level"`
}

func (d *DBServiceUsecase) CreateDBService(ctx context.Context, args *BizDBServiceArgs, currentUserUid string) (uid string, err error) {
	// 检查空间是否归档/删除
	if err := d.namespaceUsecase.isNamespaceActive(ctx, args.NamespaceUID); err != nil {
		return "", fmt.Errorf("create db service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, args.NamespaceUID); err != nil {
		return "", fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return "", fmt.Errorf("user is not namespace admin")
	}

	ds, err := newDBService(args)
	if err != nil {
		return "", fmt.Errorf("new db service failed: %w", err)
	}

	// 调用其他服务对数据源进行预检查
	if err := d.pluginUsecase.AddDBServicePreCheck(ctx, ds); err != nil {
		return "", fmt.Errorf("precheck db service failed: %w", err)
	}

	return ds.UID, d.repo.SaveDBService(ctx, ds)
}

type ListDBServicesOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      DBServiceField
	FilterBy     []pkgConst.FilterCondition
}

func (d *DBServiceUsecase) ListDBService(ctx context.Context, option *ListDBServicesOption, namespaceUid, currentUserUid string) (dbServices []*DBService, total int64, err error) {
	// 只允许系统用户查询所有数据源,同步数据到其他服务(provision)
	// 检查空间是否归档/删除
	if namespaceUid != "" {
		if err := d.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
			return nil, 0, fmt.Errorf("list db service error: %v", err)
		}
	} else if currentUserUid != pkgConst.UIDOfUserSys {
		return nil, 0, fmt.Errorf("list db service error: namespace is empty")
	}
	services, total, err := d.repo.ListDBServices(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list db services failed: %w", err)
	}
	return services, total, nil
}

func (d *DBServiceUsecase) GetActiveDBServices(ctx context.Context, dbServiceIds []string) (dbServices []*DBService, err error) {
	services, err := d.repo.GetDBServicesByIds(ctx, dbServiceIds)
	if err != nil {
		return nil, fmt.Errorf("list db services failed: %w", err)
	}

	for _, service := range services {
		if err = d.namespaceUsecase.isNamespaceActive(ctx, service.NamespaceUID); err == nil {
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
`, dbService.UID, dbService.Host, dbService.Port, dbService.AdminUser, aes.Md5(dbService.AdminPassword), dbService.AdditionalParams)
}

func (d *DBServiceUsecase) DelDBService(ctx context.Context, dbServiceUid, currentUserUid string) (err error) {
	ds, err := d.repo.GetDBService(ctx, dbServiceUid)
	if err != nil {
		return fmt.Errorf("get db service failed: %v", err)
	}
	// 检查空间是否归档/删除
	if err := d.namespaceUsecase.isNamespaceActive(ctx, ds.NamespaceUID); err != nil {
		return fmt.Errorf("delete db service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, ds.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	// 调用其他服务对数据源进行预检查
	if err := d.pluginUsecase.DelDBServicePreCheck(ctx, ds.GetUID()); err != nil {
		return fmt.Errorf("precheck del db service failed: %v", err)
	}

	if err := d.repo.DelDBService(ctx, dbServiceUid); nil != err {
		return fmt.Errorf("delete data service error: %v", err)
	}
	return nil
}

func (d *DBServiceUsecase) CheckDBServiceExist(ctx context.Context, dbServiceUids []string) (bool, error) {
	return d.repo.CheckDBServiceExist(ctx, dbServiceUids)
}

func (d *DBServiceUsecase) GetDBService(ctx context.Context, dbServiceUid string) (*DBService, error) {
	return d.repo.GetDBService(ctx, dbServiceUid)
}

func (d *DBServiceUsecase) UpdateDBService(ctx context.Context, dbServiceUid string, updateDBService *BizDBServiceArgs, currentUserUid string) (err error) {
	ds, err := d.repo.GetDBService(ctx, dbServiceUid)
	if err != nil {
		return fmt.Errorf("get db service failed: %v", err)
	}
	// 检查空间是否归档/删除
	if err := d.namespaceUsecase.isNamespaceActive(ctx, ds.NamespaceUID); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, ds.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	// check
	{
		if ds.DBType != updateDBService.DBType {
			return fmt.Errorf("update db service db type is unsupported")
		}
		source, err := pkgConst.ParseDBServiceSource(ds.Source)
		if err != nil {
			return fmt.Errorf("parse db service source failed: %v", err)
		}

		if source != pkgConst.DBServiceSourceNameDMS {
			return fmt.Errorf("external db service can not be updated")
		}
		if updateDBService.Host == "" || updateDBService.Port == "" ||
			updateDBService.AdminUser == "" || updateDBService.Business == "" {
			return fmt.Errorf("db service's host,port,user,business can't be empty")
		}
	}
	// update
	{
		if updateDBService.Desc != nil {
			ds.Desc = *updateDBService.Desc
		}
		if updateDBService.AdminPassword != nil {
			if *updateDBService.AdminPassword == "" {
				return fmt.Errorf("password can't be empty")
			}
			ds.AdminPassword = *updateDBService.AdminPassword
		}

		ds.Host = updateDBService.Host
		ds.Port = updateDBService.Port
		ds.AdminUser = updateDBService.AdminUser
		ds.Business = updateDBService.Business
		ds.AdditionalParams = updateDBService.AdditionalParams
		ds.MaintenancePeriod = updateDBService.MaintenancePeriod

		// 支持新增和更新sqleConfig，不允许删除sqle配置
		if updateDBService.SQLQueryConfig != nil && updateDBService.RuleTemplateName != "" {
			ds.SQLEConfig = &SQLEConfig{
				SQLQueryConfig:   updateDBService.SQLQueryConfig,
				RuleTemplateName: updateDBService.RuleTemplateName,
				RuleTemplateID:   updateDBService.RuleTemplateID,
			}
		}
	}

	if err := d.repo.UpdateDBService(ctx, ds); nil != err {
		return fmt.Errorf("update db service error: %v", err)
	}
	return nil
}

type IsConnectableParams struct {
	DBType           pkgConst.DBType
	Host             string
	Port             string
	User             string
	Password         string
	AdditionalParams pkgParams.Params
}

func (d *DBServiceUsecase) IsConnectable(ctx context.Context, params IsConnectableParams) (bool, error) {
	switch params.DBType {
	case pkgConst.DBTypeMySQL:
		return database.NewMysqlManager(params.Host, params.Port, params.User, params.Password).IsConnectable(ctx)
	case pkgConst.DBTypeOracle:
		return false, nil
	case pkgConst.DBTypeOceanBaseMySQL:
		return false, nil
	case pkgConst.DBTypePostgreSQL:
		return false, nil
	case pkgConst.DBTypeSQLServer:
		return false, nil
	default:
		return false, fmt.Errorf("%s does not exist", params.DBType)
	}
}
