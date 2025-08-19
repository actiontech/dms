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
	BasicUsecase                *biz.BasicUsecase
	ResourceOverviewUsecase     *biz.ResourceOverviewUsecase
	BusinessTagUsecase          *biz.BusinessTagUsecase
	PluginUsecase               *biz.PluginUsecase
	DBServiceUsecase            *biz.DBServiceUsecase
	DBServiceSyncTaskUsecase    *biz.DBServiceSyncTaskUsecase
	EnvironmentTagUsecase       *biz.EnvironmentTagUsecase
	LoginConfigurationUsecase   *biz.LoginConfigurationUsecase
	UserUsecase                 *biz.UserUsecase
	UserGroupUsecase            *biz.UserGroupUsecase
	RoleUsecase                 *biz.RoleUsecase
	OpPermissionUsecase         *biz.OpPermissionUsecase
	MemberUsecase               *biz.MemberUsecase
	MemberGroupUsecase          *biz.MemberGroupUsecase
	OpPermissionVerifyUsecase   *biz.OpPermissionVerifyUsecase
	ProjectUsecase              *biz.ProjectUsecase
	DmsProxyUsecase             *biz.DmsProxyUsecase
	Oauth2ConfigurationUsecase  *biz.Oauth2ConfigurationUsecase
	OAuth2SessionUsecase        *biz.OAuth2SessionUsecase
	LDAPConfigurationUsecase    *biz.LDAPConfigurationUsecase
	SMTPConfigurationUsecase    *biz.SMTPConfigurationUsecase
	WeChatConfigurationUsecase  *biz.WeChatConfigurationUsecase
	WebHookConfigurationUsecase *biz.WebHookConfigurationUsecase
	SmsConfigurationUseCase     *biz.SmsConfigurationUseCase
	IMConfigurationUsecase      *biz.IMConfigurationUsecase
	CompanyNoticeUsecase        *biz.CompanyNoticeUsecase
	LicenseUsecase              *biz.LicenseUsecase
	ClusterUsecase              *biz.ClusterUsecase
	DataExportWorkflowUsecase   *biz.DataExportWorkflowUsecase
	CbOperationLogUsecase       *biz.CbOperationLogUsecase
	DataMaskingUsecase          *biz.DataMaskingUsecase
	AuthAccessTokenUseCase      *biz.AuthAccessTokenUsecase
	SwaggerUseCase              *biz.SwaggerUseCase
	GatewayUsecase              *biz.GatewayUsecase
	SystemVariableUsecase       *biz.SystemVariableUsecase
	log                         *utilLog.Helper
	shutdownCallback            func() error
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
	environmentTagUsecase := biz.EnvironmentTagUsecase{}
	businessTagUsecase := biz.NewBusinessTagUsecase(storage.NewBusinessTagRepo(logger, st), logger)
	projectRepo := storage.NewProjectRepo(logger, st)
	projectUsecase := biz.NewProjectUsecase(logger, tx, projectRepo, &memberUsecase, opPermissionVerifyUsecase, pluginUseCase, businessTagUsecase, &environmentTagUsecase)
	dbServiceRepo := storage.NewDBServiceRepo(logger, st)
	dmsProxyTargetRepo := storage.NewProxyTargetRepo(logger, st)
	resourceOverviewUsecase := biz.NewResourceOverviewUsecase(logger, projectRepo, dbServiceRepo, *opPermissionVerifyUsecase, storage.NewResourceOverviewRepo(logger, st), dmsProxyTargetRepo)
	environmentTagUsecase = *biz.NewEnvironmentTagUsecase(storage.NewEnvironmentTagRepo(logger, st), logger, projectUsecase, opPermissionVerifyUsecase)
	dbServiceUseCase := biz.NewDBServiceUsecase(logger, dbServiceRepo, pluginUseCase, opPermissionVerifyUsecase, projectUsecase, dmsProxyTargetRepo, &environmentTagUsecase)
	dbServiceTaskRepo := storage.NewDBServiceSyncTaskRepo(logger, st)
	dbServiceTaskUsecase := biz.NewDBServiceSyncTaskUsecase(logger, dbServiceTaskRepo, opPermissionVerifyUsecase, projectUsecase, dbServiceUseCase, &environmentTagUsecase)
	ldapConfigurationRepo := storage.NewLDAPConfigurationRepo(logger, st)
	ldapConfigurationUsecase := biz.NewLDAPConfigurationUsecase(logger, tx, ldapConfigurationRepo)
	userRepo := storage.NewUserRepo(logger, st)
	userGroupRepo := storage.NewUserGroupRepo(logger, st)
	opPermissionRepo := storage.NewOpPermissionRepo(logger, st)
	opPermissionUsecase := biz.NewOpPermissionUsecase(logger, tx, opPermissionRepo, pluginUseCase)
	cloudbeaverRepo := storage.NewCloudbeaverRepo(logger, st)
	loginConfigurationRepo := storage.NewLoginConfigurationRepo(logger, st)
	loginConfigurationUsecase := biz.NewLoginConfigurationUsecase(logger, tx, loginConfigurationRepo)

	gatewayUsecase, err := biz.NewDmsGatewayUsecase(logger, storage.NewGatewayRepo(logger, st))
	if err != nil {
		return nil, fmt.Errorf("failed to new dms gateway use case: %v", err)
	}

	userUsecase := biz.NewUserUsecase(logger, tx, userRepo, userGroupRepo, pluginUseCase, opPermissionUsecase, opPermissionVerifyUsecase, loginConfigurationUsecase, ldapConfigurationUsecase, cloudbeaverRepo, gatewayUsecase)
	userGroupUsecase := biz.NewUserGroupUsecase(logger, tx, userGroupRepo, userRepo, pluginUseCase, opPermissionVerifyUsecase)
	roleRepo := storage.NewRoleRepo(logger, st)
	memberRepo := storage.NewMemberRepo(logger, st)
	roleUsecase := biz.NewRoleUsecase(logger, tx, roleRepo, opPermissionRepo, memberRepo, pluginUseCase, opPermissionVerifyUsecase)
	dmsConfigRepo := storage.NewDMSConfigRepo(logger, st)
	dmsConfigUsecase := biz.NewDMSConfigUseCase(logger, dmsConfigRepo)
	memberUsecase = *biz.NewMemberUsecase(logger, tx, memberRepo, userUsecase, roleUsecase, dbServiceUseCase, opPermissionVerifyUsecase, projectUsecase, pluginUseCase)
	memberGroupRepo := storage.NewMemberGroupRepo(logger, st)
	memberGroupUsecase := biz.NewMemberGroupUsecase(logger, tx, memberGroupRepo, userUsecase, roleUsecase, dbServiceUseCase, opPermissionVerifyUsecase, projectUsecase, &memberUsecase, pluginUseCase)
	dmsProxyUsecase, err := biz.NewDmsProxyUsecase(logger, dmsProxyTargetRepo, opts.APIServiceOpts, opPermissionUsecase, roleUsecase)
	if err != nil {
		return nil, fmt.Errorf("failed to new dms proxy usecase: %v", err)
	}
	oauth2SessionRepo := storage.NewOAuth2SessionRepo(logger, st)
	oauth2SessionUsecase := biz.NewOAuth2SessionUsecase(logger, tx, oauth2SessionRepo)
	oauth2ConfigurationRepo := storage.NewOauth2ConfigurationRepo(logger, st)
	oauth2ConfigurationUsecase := biz.NewOauth2ConfigurationUsecase(logger, tx, oauth2ConfigurationRepo, userUsecase, oauth2SessionUsecase)
	companyNoticeRepo := storage.NewCompanyNoticeRepo(logger, st)
	companyNoticeRepoUsecase := biz.NewCompanyNoticeUsecase(logger, tx, companyNoticeRepo, userUsecase)
	smtpConfigurationRepo := storage.NewSMTPConfigurationRepo(logger, st)
	smtpConfigurationUsecase := biz.NewSMTPConfigurationUsecase(logger, tx, smtpConfigurationRepo)
	wechatConfigurationRepo := storage.NewWeChatConfigurationRepo(logger, st)
	wechatConfigurationUsecase := biz.NewWeChatConfigurationUsecase(logger, tx, wechatConfigurationRepo)
	webhookConfigurationRepo := storage.NewWebHookConfigurationRepo(logger, st)
	webhookConfigurationUsecase := biz.NewWebHookConfigurationUsecase(logger, tx, webhookConfigurationRepo)
	smsConfigurationRepo := storage.NewSmsConfigurationRepo(logger, st)
	smsConfigurationUsecase := biz.NewSmsConfigurationUsecase(logger, tx, smsConfigurationRepo, userUsecase)
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

	cronTask := biz.NewCronTaskUsecase(logger, DataExportWorkflowUsecase, CbOperationLogUsecase, oauth2SessionUsecase)
	err = cronTask.InitialTask()
	if err != nil {
		return nil, fmt.Errorf("failed to new cron task: %v", err)
	}

	s := &DMSService{
		BasicUsecase:                basicUsecase,
		ResourceOverviewUsecase:     resourceOverviewUsecase,
		BusinessTagUsecase:          businessTagUsecase,
		EnvironmentTagUsecase:       &environmentTagUsecase,
		PluginUsecase:               pluginUseCase,
		DBServiceUsecase:            dbServiceUseCase,
		DBServiceSyncTaskUsecase:    dbServiceTaskUsecase,
		LoginConfigurationUsecase:   loginConfigurationUsecase,
		UserUsecase:                 userUsecase,
		UserGroupUsecase:            userGroupUsecase,
		RoleUsecase:                 roleUsecase,
		OpPermissionUsecase:         opPermissionUsecase,
		MemberUsecase:               &memberUsecase,
		MemberGroupUsecase:          memberGroupUsecase,
		OpPermissionVerifyUsecase:   opPermissionVerifyUsecase,
		ProjectUsecase:              projectUsecase,
		DmsProxyUsecase:             dmsProxyUsecase,
		Oauth2ConfigurationUsecase:  oauth2ConfigurationUsecase,
		OAuth2SessionUsecase:        oauth2SessionUsecase,
		LDAPConfigurationUsecase:    ldapConfigurationUsecase,
		SMTPConfigurationUsecase:    smtpConfigurationUsecase,
		WeChatConfigurationUsecase:  wechatConfigurationUsecase,
		WebHookConfigurationUsecase: webhookConfigurationUsecase,
		IMConfigurationUsecase:      imConfigurationUsecase,
		SmsConfigurationUseCase:     smsConfigurationUsecase,
		CompanyNoticeUsecase:        companyNoticeRepoUsecase,
		LicenseUsecase:              LicenseUsecase,
		ClusterUsecase:              clusterUsecase,
		DataExportWorkflowUsecase:   DataExportWorkflowUsecase,
		CbOperationLogUsecase:       CbOperationLogUsecase,
		DataMaskingUsecase:          dataMaskingUsecase,
		AuthAccessTokenUseCase:      authAccessTokenUsecase,
		SwaggerUseCase:              swaggerUseCase,
		GatewayUsecase:              gatewayUsecase,
		SystemVariableUsecase:       biz.NewSystemVariableUsecase(logger, storage.NewSystemVariableRepo(logger, st)),
		log:                         utilLog.NewHelper(logger, utilLog.WithMessageKey("dms.service")),
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
