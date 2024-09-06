package locale

import (
	"embed"

	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
)

//go:embed active.*.toml
var localeFS embed.FS

var Bundle *i18nPkg.Bundle

func MustInit(l i18nPkg.Log) {
	b, err := i18nPkg.NewBundleFromTomlDir(localeFS, l)
	if err != nil {
		panic(err)
	}
	Bundle = b
}
