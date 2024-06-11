package service

import (
	"github.com/actiontech/dms/api"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	"github.com/labstack/echo/v4"
)

func (d *DMSService) RegisterSwagger(c echo.Context) error {
	_, ok := api.GetSwaggerDoc(api.SqlSwaggerTypeKey)
	if !ok {
		reply, err := d.GetSqlSwaggerContent(c)
		if err != nil {
			return err
		}
		api.RegisterSwaggerDoc(api.SqlSwaggerTypeKey, reply)
	}

	return nil
}

func (d *DMSService) GetSqlSwaggerContent(c echo.Context) ([]byte, error) {
	target, err := d.DmsProxyUsecase.GetTargetByName(c.Request().Context(), cloudbeaver.SQLEProxyName)
	if err != nil {
		return nil, err
	}

	url := target.URL.String() + "/swagger_file"

	header := map[string]string{
		echo.HeaderAuthorization: pkgHttp.DefaultDMSToken,
	}

	resp := struct {
		Content []byte `json:"content"`
	}{}

	err = pkgHttp.Get(c.Request().Context(), url, header, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}
