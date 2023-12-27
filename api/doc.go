package api

import (
	"embed"
	"io"
	"path"

	_ "embed"
	"io/fs"

	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/swaggo/swag"
)

//go:embed docs
var multi embed.FS

type SwaggerDoc struct {
	file fs.File
}

func (sd *SwaggerDoc) ReadDoc() string {
	doc, err := io.ReadAll(sd.file)
	if err != nil {
		return ""
	}
	return string(doc)
}

func init() {
	docDir, err := multi.ReadDir("docs")
	if err != nil {
		return
	}
	// - docs
	//   - dms
	//		swagger.yaml
	//	 - sqle
	//      swagger.yaml
	for _, fd := range docDir {
		// dms or sqle
		if fd.IsDir() {
			currentFdName := fd.Name()
			files, err := fs.Sub(multi, path.Join("docs", currentFdName))
			if err != nil {
				return
			}

			swagFile, err := files.Open("swagger.yaml")
			if err != nil {
				return
			}

			url := path.Join(currentFdName, "doc.yaml")
			swag.Register(url, &SwaggerDoc{file: swagFile})
			ConfigFunc = append(ConfigFunc, func(config *echoSwagger.Config) {
				config.URLs = append(config.URLs, url)
				config.InstanceName = url
			})
		}
	}
}

var Config *echoSwagger.Config

var ConfigFunc []func(*echoSwagger.Config) = []func(*echoSwagger.Config){
	func(config *echoSwagger.Config) {
		config.URLs = []string{}
		Config = config
	},
}
