package service

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/apiserver/conf"
	maskingBiz "github.com/actiontech/dms/internal/data_masking/biz"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type DMSService struct {
	BasicUsecase                 *biz.BasicUsecase
	PluginUsecase                *biz.PluginUsecase
	DBServiceUsecase             *biz.DBServiceUsecase
	DBServiceSyncTaskUsecase *biz.DBServiceSyncTaskUsecase
	UserUsecase                  *biz.UserUsecase
	UserGroupUsecase             *biz.UserGroupUsecase
	RoleUsecase                  *biz.RoleUsecase
	OpPermissionUsecase          *biz.OpPermissionUsecase
	MemberUsecase                *biz.MemberUsecase
	MemberGroupUsecase           *biz.MemberGroupUsecase
	OpPermissionVerifyUsecase    *biz.OpPermissionVerifyUsecase
	ProjectUsecase               *biz.ProjectUsecase
	DmsProxyUsecase              *biz.DmsProxyUsecase
	Oauth2ConfigurationUsecase   *biz.Oauth2ConfigurationUsecase
	LDAPConfigurationUsecase     *biz.LDAPConfigurationUsecase
	SMTPConfigurationUsecase     *biz.SMTPConfigurationUsecase
	WeChatConfigurationUsecase   *biz.WeChatConfigurationUsecase
	WebHookConfigurationUsecase  *biz.WebHookConfigurationUsecase
	IMConfigurationUsecase       *biz.IMConfigurationUsecase
	CompanyNoticeUsecase         *biz.CompanyNoticeUsecase
	LicenseUsecase               *biz.LicenseUsecase
	ClusterUsecase               *biz.ClusterUsecase
	DataExportWorkflowUsecase    *biz.DataExportWorkflowUsecase
	CbOperationLogUsecase        *biz.CbOperationLogUsecase
	DataMaskingUsecase           *biz.DataMaskingUsecase
	AuthAccessTokenUseCase       *biz.AuthAccessTokenUsecase
	SwaggerUseCase               *biz.SwaggerUseCase
	log                          *utilLog.Helper
	shutdownCallback             func() error
}

func NewAndInitDMSService(logger utilLog.Logger, opts *conf.DMSOptions) (*DMSService, error) {
	st, err := storage.NewStorage(logger, &storage.StorageConfig{
		User:        opts.ServiceOpts.Database.UserName,
		Password:    opts.ServiceOpts.Database.Password,
		Host:        opts.ServiceOpts.Database.Host,
		Port:        opts.ServiceOpts.Database.Port,
		Schema:      opts.ServiceOpts.Database.Database,
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
	memberUsecase := biz.MemberUsecase{}

	projectRepo := storage.NewProjectRepo(logger, st)
	projectUsecase := biz.NewProjectUsecase(logger, tx, projectRepo, &memberUsecase, opPermissionVerifyUsecase, pluginUseCase)
	dbServiceRepo := storage.NewDBServiceRepo(logger, st)
	dmsProxyTargetRepo := storage.NewProxyTargetRepo(logger, st)
	dbServiceUseCase := biz.NewDBServiceUsecase(logger, dbServiceRepo, pluginUseCase, opPermissionVerifyUsecase, projectUsecase, dmsProxyTargetRepo, opts.DatabaseDriverOptions)
	databaseSourceServiceRepo := storage.NewDBServiceSyncTaskRepo(logger, st)
	databaseSourceServiceUsecase := biz.NewDBServiceSyncTaskUsecase(logger, databaseSourceServiceRepo, opPermissionVerifyUsecase, projectUsecase, dbServiceUseCase)
	ldapConfigurationRepo := storage.NewLDAPConfigurationRepo(logger, st)
	ldapConfigurationUsecase := biz.NewLDAPConfigurationUsecase(logger, tx, ldapConfigurationRepo)
	userRepo := storage.NewUserRepo(logger, st)
	userGroupRepo := storage.NewUserGroupRepo(logger, st)
	opPermissionRepo := storage.NewOpPermissionRepo(logger, st)
	opPermissionUsecase := biz.NewOpPermissionUsecase(logger, tx, opPermissionRepo, pluginUseCase)
	userUsecase := biz.NewUserUsecase(logger, tx, userRepo, userGroupRepo, pluginUseCase, opPermissionUsecase, opPermissionVerifyUsecase, ldapConfigurationUsecase)
	userGroupUsecase := biz.NewUserGroupUsecase(logger, tx, userGroupRepo, userRepo, pluginUseCase, opPermissionVerifyUsecase)
	roleRepo := storage.NewRoleRepo(logger, st)
	memberRepo := storage.NewMemberRepo(logger, st)
	roleUsecase := biz.NewRoleUsecase(logger, tx, roleRepo, opPermissionRepo, memberRepo, pluginUseCase, opPermissionVerifyUsecase)
	dmsConfigRepo := storage.NewDMSConfigRepo(logger, st)
	dmsConfigUsecase := biz.NewDMSConfigUseCase(logger, dmsConfigRepo)
	memberUsecase = *biz.NewMemberUsecase(logger, tx, memberRepo, userUsecase, roleUsecase, dbServiceUseCase, opPermissionVerifyUsecase, projectUsecase)
	memberGroupRepo := storage.NewMemberGroupRepo(logger, st)
	memberGroupUsecase := biz.NewMemberGroupUsecase(logger, tx, memberGroupRepo, userUsecase, roleUsecase, dbServiceUseCase, opPermissionVerifyUsecase, projectUsecase, &memberUsecase)
	dmsProxyUsecase, err := biz.NewDmsProxyUsecase(logger, dmsProxyTargetRepo, opts.APIServiceOpts.Port)
	oauth2ConfigurationRepo := storage.NewOauth2ConfigurationRepo(logger, st)
	oauth2ConfigurationUsecase := biz.NewOauth2ConfigurationUsecase(logger, tx, oauth2ConfigurationRepo, userUsecase)
	companyNoticeRepo := storage.NewCompanyNoticeRepo(logger, st)
	companyNoticeRepoUsecase := biz.NewCompanyNoticeUsecase(logger, tx, companyNoticeRepo, userUsecase)
	smtpConfigurationRepo := storage.NewSMTPConfigurationRepo(logger, st)
	smtpConfigurationUsecase := biz.NewSMTPConfigurationUsecase(logger, tx, smtpConfigurationRepo)
	wechatConfigurationRepo := storage.NewWeChatConfigurationRepo(logger, st)
	wechatConfigurationUsecase := biz.NewWeChatConfigurationUsecase(logger, tx, wechatConfigurationRepo)
	webhookConfigurationRepo := storage.NewWebHookConfigurationRepo(logger, st)
	webhookConfigurationUsecase := biz.NewWebHookConfigurationUsecase(logger, tx, webhookConfigurationRepo)
	imConfigurationRepo := storage.NewIMConfigurationRepo(logger, st)
	imConfigurationUsecase := biz.NewIMConfigurationUsecase(logger, tx, imConfigurationRepo)
	basicConfigRepo := storage.NewBasicConfigRepo(logger, st)
	basicUsecase := biz.NewBasicInfoUsecase(logger, dmsProxyUsecase, basicConfigRepo)
	clusterRepo := storage.NewClusterRepo(logger, st)
	clusterUsecase := biz.NewClusterUsecase(logger, tx, clusterRepo)
	licenseRepo := storage.NewLicenseRepo(logger, st)
	LicenseUsecase := biz.NewLicenseUsecase(logger, tx, licenseRepo, userUsecase, dbServiceUseCase, clusterUsecase)
	dataExportTaskRepo := storage.NewDataExportTaskRepo(logger, st)

	swaggerUseCase := biz.NewSwaggerUseCase(logger, dmsProxyUsecase)

	cbOperationRepo := storage.NewCbOperationLogRepo(logger, st)
	CbOperationLogUsecase := biz.NewCbOperationLogUsecase(logger, cbOperationRepo, opPermissionVerifyUsecase, dmsProxyTargetRepo)
	workflowRepo := storage.NewWorkflowRepo(logger, st)
	DataExportWorkflowUsecase := biz.NewDataExportWorkflowUsecase(logger, tx, workflowRepo, dataExportTaskRepo, dbServiceRepo, opPermissionVerifyUsecase, projectUsecase, dmsProxyTargetRepo, clusterUsecase, webhookConfigurationUsecase, userUsecase, fmt.Sprintf("%s:%d", opts.ReportHost, opts.APIServiceOpts.Port))
	dataMasking, err := maskingBiz.NewDataMaskingUseCase(logger)
	authAccessTokenUsecase := biz.NewAuthAccessTokenUsecase(logger, userUsecase)
	if err != nil {
		return nil, fmt.Errorf("failed to new data masking use case: %v", err)
	}
	dataMaskingUsecase := biz.NewMaskingUsecase(logger, dataMasking)

	cronTask := biz.NewCronTaskUsecase(logger, DataExportWorkflowUsecase, CbOperationLogUsecase)
	err = cronTask.InitialTask()
	if err != nil {
		return nil, fmt.Errorf("failed to new cron task: %v", err)
	}

	s := &DMSService{
		BasicUsecase:                 basicUsecase,
		PluginUsecase:                pluginUseCase,
		DBServiceUsecase:             dbServiceUseCase,
		DBServiceSyncTaskUsecase: databaseSourceServiceUsecase,
		UserUsecase:                  userUsecase,
		UserGroupUsecase:             userGroupUsecase,
		RoleUsecase:                  roleUsecase,
		OpPermissionUsecase:          opPermissionUsecase,
		MemberUsecase:                &memberUsecase,
		MemberGroupUsecase:           memberGroupUsecase,
		OpPermissionVerifyUsecase:    opPermissionVerifyUsecase,
		ProjectUsecase:               projectUsecase,
		DmsProxyUsecase:              dmsProxyUsecase,
		Oauth2ConfigurationUsecase:   oauth2ConfigurationUsecase,
		LDAPConfigurationUsecase:     ldapConfigurationUsecase,
		SMTPConfigurationUsecase:     smtpConfigurationUsecase,
		WeChatConfigurationUsecase:   wechatConfigurationUsecase,
		WebHookConfigurationUsecase:  webhookConfigurationUsecase,
		IMConfigurationUsecase:       imConfigurationUsecase,
		CompanyNoticeUsecase:         companyNoticeRepoUsecase,
		LicenseUsecase:               LicenseUsecase,
		ClusterUsecase:               clusterUsecase,
		DataExportWorkflowUsecase:    DataExportWorkflowUsecase,
		CbOperationLogUsecase:        CbOperationLogUsecase,
		DataMaskingUsecase:           dataMaskingUsecase,
		AuthAccessTokenUseCase:       authAccessTokenUsecase,
		SwaggerUseCase:               swaggerUseCase,
		log:                          utilLog.NewHelper(logger, utilLog.WithMessageKey("dms.service")),
		shutdownCallback: func() error {
			if err := st.Close(); nil != err {
				return fmt.Errorf("failed to close storage: %v", err)
			}
			return nil
		},
	}

	// init notification
	biz.Init(smtpConfigurationUsecase, wechatConfigurationUsecase, imConfigurationUsecase)
	// init env
	if err := biz.EnvPrepare(context.TODO(), logger, tx, dmsConfigUsecase, opPermissionUsecase, userUsecase, roleUsecase, projectUsecase); nil != err {
		return nil, fmt.Errorf("failed to prepare env: %v", err)
	}
	s.log.Debug("env prepared")

	return s, nil
}

func (a *DMSService) Shutdown() error {
	if nil != a.shutdownCallback {
		return a.shutdownCallback()
	}
	return nil
}
