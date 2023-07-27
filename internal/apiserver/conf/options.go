package conf

import (
	"fmt"

	dmsConf "github.com/actiontech/dms/internal/dms/conf"

	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type Options struct {
	APIServiceOpts  *APIServerOpts   `yaml:"api"`
	DMSServiceOpts  *dmsConf.Options `yaml:"dms"`
	CloudbeaverOpts *CloudbeaverOpts `yaml:"cloudbeaver"`
	NodeOpts        *NodeOpts        `yaml:"node"`
}

type APIServerOpts struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required"`
		Port int    `yaml:"port" validate:"required"`
	} `yaml:"http"`
}

type CloudbeaverOpts struct {
	EnableHttps   bool   `yaml:"enable_https"`
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	AdminUser     string `yaml:"admin_user"`
	AdminPassword string `yaml:"admin_password"`
}

type NodeOpts struct {
	NodeNo int64 `yaml:"nodeno" validate:"required"`
}

func ReadOptions(log utilLog.Logger, path string) (*Options, error) {
	var opts Options
	if err := utilConf.ParseYamlFile(log, path, &opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

func (o *Options) GetAPIServer() *APIServerOpts {
	return o.APIServiceOpts
}

func (api *APIServerOpts) GetHTTPAddr() string {
	return fmt.Sprintf("%v:%v", api.HTTP.Addr, api.HTTP.Port)
}
