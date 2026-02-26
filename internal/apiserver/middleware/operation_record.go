package middleware

import (
	"strings"

	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	"github.com/labstack/echo/v4"
)

type ApiInterfaceInfo struct {
	RouterPath               string
	Method                   string
	OperationType            string
	OperationAction          string
	GetProjectAndContentFunc func(c echo.Context, dms interface{}) (projectName string, content i18nPkg.I18nStr, err error)
}

var ApiInterfaceInfoList []ApiInterfaceInfo

func pathMatch(pattern, path string) bool {
	ps := strings.Split(strings.Trim(pattern, "/"), "/")
	pathSegs := strings.Split(strings.Trim(path, "/"), "/")
	if len(ps) != len(pathSegs) {
		return false
	}
	for i := range ps {
		if len(ps[i]) > 0 && ps[i][0] == ':' {
			continue
		}
		if ps[i] != pathSegs[i] {
			return false
		}
	}
	return true
}
