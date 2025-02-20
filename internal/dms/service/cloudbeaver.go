package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	maskingBiz "github.com/actiontech/dms/internal/data_masking/biz"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type CloudbeaverService struct {
	CloudbeaverUsecase *biz.CloudbeaverUsecase
	ProxyUsecase       *biz.CloudbeaverProxyUsecase
	log                *utilLog.Helper
}

func NewAndInitCloudbeaverService(logger utilLog.Logger, opts *conf.DMSOptions) (*CloudbeaverService, error) {
	// todo: because cloudbeaver required userUsecase, optimisation may be needed here
	st, err := storage.NewStorage(logger, &storage.StorageConfig{
		User:        opts.ServiceOpts.Database.UserName,
		Password:    opts.ServiceOpts.Database.Password,
		Host:        opts.ServiceOpts.Database.Host,
		Port:        opts.ServiceOpts.Database.Port,
		Schema:      opts.ServiceOpts.Database.Database,
		Debug:       opts.ServiceOpts.Database.Debug,
		AutoMigrate: opts.ServiceOpts.Database.AutoMigrate,
	})
	if nil != err {
		return nil, fmt.Errorf("failed to new data: %v", err)
	}

	tx := storage.NewTXGenerator()
	opPermissionVerifyRepo := storage.NewOpPermissionVerifyRepo(logger, st)
	opPermissionVerifyUsecase := biz.NewOpPermissionVerifyUsecase(logger, tx, opPermissionVerifyRepo)
	pluginRepo := storage.NewPluginRepo(logger, st)
	pluginUseCase, err := biz.NewDMSPluginUsecase(logger, pluginRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to new dms plugin usecase: %v", err)
	}
	// 预定义解决usecase循环依赖问题
	memberUsecase := &biz.MemberUsecase{}
	dmsProxyTargetRepo := storage.NewProxyTargetRepo(logger, st)
	projectRepo := storage.NewProjectRepo(logger, st)
	projectUsecase := biz.NewProjectUsecase(logger, tx, projectRepo, memberUsecase, opPermissionVerifyUsecase, pluginUseCase)
	dbServiceRepo := storage.NewDBServiceRepo(logger, st)
	dbServiceUseCase := biz.NewDBServiceUsecase(logger, dbServiceRepo, pluginUseCase, opPermissionVerifyUsecase, projectUsecase, dmsProxyTargetRepo)

	ldapConfigurationRepo := storage.NewLDAPConfigurationRepo(logger, st)
	ldapConfigurationUsecase := biz.NewLDAPConfigurationUsecase(logger, tx, ldapConfigurationRepo)
	userRepo := storage.NewUserRepo(logger, st)
	userGroupRepo := storage.NewUserGroupRepo(logger, st)
	opPermissionRepo := storage.NewOpPermissionRepo(logger, st)
	opPermissionUsecase := biz.NewOpPermissionUsecase(logger, tx, opPermissionRepo, pluginUseCase)
	cloudbeaverRepo := storage.NewCloudbeaverRepo(logger, st)
	loginConfigurationRepo := storage.NewLoginConfigurationRepo(logger, st)
	loginConfigurationUsecase := biz.NewLoginConfigurationUsecase(logger, tx, loginConfigurationRepo)
	userUsecase := biz.NewUserUsecase(logger, tx, userRepo, userGroupRepo, pluginUseCase, opPermissionUsecase, opPermissionVerifyUsecase, loginConfigurationUsecase, ldapConfigurationUsecase, cloudbeaverRepo)
	dataMasking, err := maskingBiz.NewDataMaskingUseCase(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to new data masking use case: %v", err)
	}
	dataMaskingUsecase := biz.NewMaskingUsecase(logger, dataMasking)
	dmsConfigRepo := storage.NewDMSConfigRepo(logger, st)
	dmsConfigUseCase := biz.NewDMSConfigUseCase(logger, dmsConfigRepo)
	cbOperationLogUsecase := biz.NewCbOperationLogUsecase(logger, storage.NewCbOperationLogRepo(logger, st), opPermissionVerifyUsecase, dmsProxyTargetRepo)

	var cfg *biz.CloudbeaverCfg
	if opts.CloudbeaverOpts != nil {
		cfg = &biz.CloudbeaverCfg{
			EnableHttps:   opts.CloudbeaverOpts.EnableHttps,
			Host:          opts.CloudbeaverOpts.Host,
			Port:          opts.CloudbeaverOpts.Port,
			AdminUser:     opts.CloudbeaverOpts.AdminUser,
			AdminPassword: opts.CloudbeaverOpts.AdminPassword,
		}
	}


	cloudbeaverUsecase := biz.NewCloudbeaverUsecase(logger, cfg, userUsecase, dbServiceUseCase, opPermissionVerifyUsecase, dmsConfigUseCase, dataMaskingUsecase, cloudbeaverRepo, dmsProxyTargetRepo, cbOperationLogUsecase, projectUsecase)
	proxyUsecase := biz.NewCloudbeaverProxyUsecase(logger, cloudbeaverUsecase)

	return &CloudbeaverService{
		CloudbeaverUsecase: cloudbeaverUsecase,
		ProxyUsecase:       proxyUsecase,
		log:                utilLog.NewHelper(logger, utilLog.WithMessageKey("cloudbeaver.service")),
	}, nil
}

func (cs *CloudbeaverService) GetCloudbeaverConfiguration(ctx context.Context) (reply *dmsV1.GetSQLQueryConfigurationReply, err error) {
	cs.log.Infof("GetCloudbeaverConfiguration")
	defer func() {
		cs.log.Infof("GetCloudbeaverConfiguration; reply=%v, error=%v", reply, err)
	}()

	return &dmsV1.GetSQLQueryConfigurationReply{
		Data: struct {
			EnableSQLQuery  bool   `json:"enable_sql_query"`
			SQLQueryRootURI string `json:"sql_query_root_uri"`
		}{
			EnableSQLQuery:  cs.CloudbeaverUsecase.IsCloudbeaverConfigured(),
			SQLQueryRootURI: cs.CloudbeaverUsecase.GetRootUri(),
		},
	}, nil
}

func (cs *CloudbeaverService) Logout(session string) {
	cs.CloudbeaverUsecase.UnbindCBSession(session)
}
