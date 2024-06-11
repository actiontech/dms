package api

import (
	_ "embed"
	"fmt"
	"path"
	"sync"

	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/swaggo/swag"
)

var (
	swaggerMu   = new(sync.RWMutex)
	swaggerList = make(map[SwaggerType]*SwaggerDoc)

	//go:embed swagger.yaml
	dmsSwagYaml []byte
)

func init() {
	RegisterSwaggerDoc(DmsSwaggerTykeKey, dmsSwagYaml)
}

type SwaggerType string

const (
	SqlSwaggerTypeKey SwaggerType = "sqle"
	DmsSwaggerTykeKey SwaggerType = "dms"
)

type SwaggerDoc struct {
	file []byte
}

func (sd *SwaggerDoc) ReadDoc() string {
	return string(sd.file)
}

var ConfigList = []func(*echoSwagger.Config){
	func(config *echoSwagger.Config) {
		// for clear the default URLs
		config.URLs = []string{}
	},
}

// RegisterSwaggerDoc registers a Swagger document and adds a config function
func RegisterSwaggerDoc(swaggerType SwaggerType, file []byte) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()

	swaggerList[swaggerType] = &SwaggerDoc{file: file}

	url := swaggerType.GetUrlPath()

	swag.Register(url, &SwaggerDoc{file: file})
	ConfigList = append(ConfigList, func(config *echoSwagger.Config) {
		config.URLs = append(config.URLs, url)
		config.InstanceName = url
	})
}

// GetSwaggerDoc returns the Swagger document by the given type
func GetSwaggerDoc(swaggerType SwaggerType) (*SwaggerDoc, bool) {
	swaggerMu.RLock()
	defer swaggerMu.RUnlock()

	doc, ok := swaggerList[swaggerType]

	return doc, ok
}

// GetAllSwaggerDocs returns all the registered Swagger documents
func GetAllSwaggerDocs() map[SwaggerType]*SwaggerDoc {
	swaggerMu.RLock()
	defer swaggerMu.RUnlock()

	return swaggerList
}

func (t SwaggerType) GetUrlPath() string {
	return path.Join(fmt.Sprintf("%v", t), "doc.yaml")
}
