package api

import (
	_ "embed"
	"fmt"
	"path"
	"sync"

	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
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
	SqleSwaggerTypeKey SwaggerType = _const.SqleComponentName
	DmsSwaggerTykeKey  SwaggerType = _const.DmsComponentName
)

type SwaggerDoc struct {
	file []byte
}

func (sd *SwaggerDoc) ReadDoc() string {
	return string(sd.file)
}

// RegisterSwaggerDoc registers a Swagger document and adds a config function
func RegisterSwaggerDoc(swaggerType SwaggerType, file []byte) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()

	swaggerList[swaggerType] = &SwaggerDoc{file: file}

	instanceName := swaggerType.GetUrlPath()

	swag.Register(instanceName, &SwaggerDoc{file: file})
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
