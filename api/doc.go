package api

import (
	_ "embed"
	"path"

	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/swaggo/swag"
)

type SwaggerDoc struct {
	file []byte
}

func (sd *SwaggerDoc) ReadDoc() string {
	return string(sd.file)
}

//go:embed docs/dms/swagger.yaml
var dmsSwagYaml []byte

var ConfigList = []func(*echoSwagger.Config){
	func(config *echoSwagger.Config) {
		// 为了将echo-swagger默认config的URLs置为空
		config.URLs = []string{}
	},
}

func init() {
	RegisterSwaggerDoc("dms", dmsSwagYaml)
}

// RegisterSwaggerDoc registers a Swagger document and adds a config function
func RegisterSwaggerDoc(basePath string, file []byte) {
	url := path.Join(basePath, "doc.yaml")
	swag.Register(url, &SwaggerDoc{file: file})
	ConfigList = append(ConfigList, func(config *echoSwagger.Config) {
		config.URLs = append(config.URLs, url)
		config.InstanceName = url
	})
}
