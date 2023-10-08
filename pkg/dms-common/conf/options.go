package conf

import "fmt"

type BaseOptions struct {
	ID             int64          `yaml:"id" validate:"required"`
	APIServiceOpts *APIServerOpts `yaml:"api"`
	SecretKey      string         `yaml:"secret_key"`
}

type APIServerOpts struct {
	Addr string `yaml:"addr" validate:"required"`
	Port int    `yaml:"port" validate:"required"`
}

func (o *BaseOptions) GetAPIServer() *APIServerOpts {
	return o.APIServiceOpts
}

func (api *APIServerOpts) GetHTTPAddr() string {
	return fmt.Sprintf("%v:%v", api.Addr, api.Port)
}
