package biz

import (
	"net/http"
	"time"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

const (
	OdcRootUri          = "/odc_query"
)

// OdcCfg ODC配置结构
type OdcCfg struct {
	EnableHttps   bool   `yaml:"enable_https"`
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	AdminUser     string `yaml:"admin_user"`
	AdminPassword string `yaml:"admin_password"`
	APIKey        string `yaml:"api_key"`       // ODC API密钥
	ClientID      string `yaml:"client_id"`     // ODC客户端ID
	ClientSecret  string `yaml:"client_secret"` // ODC客户端密钥
}

// OdcUser ODC用户映射
type OdcUser struct {
	DMSUserID      string    `json:"dms_user_id"`
	DMSFingerprint string    `json:"dms_fingerprint"`
	OdcUserID      string    `json:"odc_user_id"`
	OdcSessionID   string    `json:"odc_session_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OdcConnection ODC连接映射
type OdcConnection struct {
	DMSDBServiceID          string    `json:"dms_db_service_id"`
	Purpose                 string    `json:"purpose"`
	DMSUserId               string    `json:"dms_user_id"`
	DMSDBServiceFingerprint string    `json:"dms_db_service_fingerprint"`
	OdcConnectionID         string    `json:"odc_connection_id"`
	OdcProjectID            string    `json:"odc_project_id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// OdcRepo ODC存储接口
type OdcRepo interface {

}

// OdcUsecase ODC业务逻辑层
type OdcUsecase struct {
	odcCfg                    *OdcCfg
	log                       *utilLog.Helper
	userUsecase               *UserUsecase
	dbServiceUsecase          *DBServiceUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	dmsConfigUseCase          *DMSConfigUseCase
	dataMaskingUseCase        *DataMaskingUsecase
	cbOperationLogUsecase     *CbOperationLogUsecase
	projectUsecase            *ProjectUsecase
	repo                      OdcRepo
	proxyTargetRepo           ProxyTargetRepo
	httpClient                *http.Client
}

// NewOdcUsecase 创建ODC业务逻辑实例
func NewOdcUsecase(log utilLog.Logger, cfg *OdcCfg, userUsecase *UserUsecase, dbServiceUsecase *DBServiceUsecase, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, dmsConfigUseCase *DMSConfigUseCase, dataMaskingUseCase *DataMaskingUsecase, odcRepo OdcRepo, proxyTargetRepo ProxyTargetRepo, cbOperationLogUsecase *CbOperationLogUsecase, projectUsecase *ProjectUsecase) *OdcUsecase {
	return &OdcUsecase{
		repo:                      odcRepo,
		proxyTargetRepo:           proxyTargetRepo,
		userUsecase:               userUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		dmsConfigUseCase:          dmsConfigUseCase,
		dataMaskingUseCase:        dataMaskingUseCase,
		cbOperationLogUsecase:     cbOperationLogUsecase,
		projectUsecase:            projectUsecase,
		odcCfg:                    cfg,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.odc")),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsOdcConfigured 检查ODC是否已配置
func (o *OdcUsecase) IsOdcConfigured() bool {
	if o.odcCfg == nil {
		return false
	}
	return o.odcCfg != nil && o.odcCfg.Host != "" && o.odcCfg.Port != ""
}

// GetRootUri 获取ODC根URI
func (cu *OdcUsecase) GetRootUri() string {
	return OdcRootUri
}

// UnbindOdcSession 解绑ODC会话
func (o *OdcUsecase) UnbindOdcSession(sessionID string) {
	// 实现会话解绑逻辑
	o.log.Infof("unbinding ODC session: %s", sessionID)
}
