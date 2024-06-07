package service

import (
	"context"
	"time"

	"github.com/actiontech/dms/api"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	"github.com/labstack/echo/v4"
)

var key = "k"
var m = make(map[string][]byte)

type TT struct {
	Content []byte
}

func (d *DMSService) RegisterSwagger(c echo.Context) error {
	reply, ok := m[key]
	if !ok {
		target, err := d.DmsProxyUsecase.GetTargetByName(c.Request().Context(), cloudbeaver.SQLEProxyName)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		url := target.URL.String() + "/swagger_file"

		header := map[string]string{
			"Authorization": pkgHttp.DefaultDMSToken,
		}

		resp := &TT{}
		err = pkgHttp.Get(ctx, url, header, nil, &resp)
		if err != nil {
			return err
		}

		reply = resp.Content
		m[key] = reply

		api.RegisterSwaggerDoc("sqle", reply)
	}

	return nil
}
