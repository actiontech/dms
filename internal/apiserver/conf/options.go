package conf

import (
	workbench "github.com/actiontech/dms/internal/sql_workbench/config"
	dmsCommonConf "github.com/actiontech/dms/pkg/dms-common/conf"
	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type Options struct {
	DMS  DMSOptions  `yaml:"dms" validate:"required"`
	SQLE SQLEOptions `yaml:"sqle"`
}

type SQLEOptions struct {
	OptimizationConfig struct {
		OptimizationKey string `yaml:"optimization_key"`
		OptimizationUrl string `yaml:"optimization_url"`
	} `yaml:"optimization_config"`
}

type DMSOptions struct {
	dmsCommonConf.BaseOptions `yaml:",inline"`
	CloudbeaverOpts           *CloudbeaverOpts `yaml:"cloudbeaver"`
	SqlWorkBenchOpts          *workbench.SqlWorkbenchOpts `yaml:"sql_workbench"`
	ServiceOpts               *ServiceOptions  `yaml:"service"`
}

type CloudbeaverOpts struct {
	EnableHttps   bool   `yaml:"enable_https"`
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	AdminUser     string `yaml:"admin_user"`
	AdminPassword string `yaml:"admin_password"`
}

type ServiceOptions struct {
	Database struct {
		UserName    string `yaml:"username" `
		Password    string `yaml:"password" `
		Host        string `yaml:"host" validate:"required"`
		Port        string `yaml:"port" validate:"required"`
		Database    string `yaml:"database" validate:"required"`
		AutoMigrate bool   `yaml:"auto_migrate"`
		Debug       bool   `yaml:"debug"`
	} `yaml:"database"`
	Log struct {
		Level           string `yaml:"level"`
		Path            string `yaml:"path"`
		MaxSizeMB       int    `yaml:"max_size_mb"`
		MaxBackupNumber int    `yaml:"max_backup_number"`
	} `yaml:"log"`
}

var optimizationEnabled bool

func IsOptimizationEnabled() bool {
	return optimizationEnabled
}

func ReadOptions(log utilLog.Logger, path string) (*DMSOptions, error) {
	var opts Options
	if err := utilConf.ParseYamlFile(log, path, &opts); err != nil {
		return nil, err
	}
	optimizationEnabled = getOptimizationEnabled(&opts)
	return &opts.DMS, nil
}
