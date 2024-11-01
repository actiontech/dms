package conf

import (
	dmsCommonConf "github.com/actiontech/dms/pkg/dms-common/conf"
	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgParams "github.com/actiontech/dms/pkg/params"
)

var IsOptimizationEnabled bool

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
	CloudbeaverOpts           *CloudbeaverOpts       `yaml:"cloudbeaver"`
	ServiceOpts               *ServiceOptions        `yaml:"service"`
	DatabaseDriverOptions     []DatabaseDriverOption `yaml:"database_driver_options"`
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

type DatabaseDriverOption struct {
	DbType   string           `yaml:"db_type"`
	LogoPath string           `yaml:"logo_path"`
	Params   pkgParams.Params `yaml:"params"`
}

func ReadOptions(log utilLog.Logger, path string) (*DMSOptions, error) {
	var opts Options
	if err := utilConf.ParseYamlFile(log, path, &opts); err != nil {
		return nil, err
	}
	IsOptimizationEnabled = getOptimizationEnabled(&opts)
	return &opts.DMS, nil
}
