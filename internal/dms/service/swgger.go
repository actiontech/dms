package service

import (
	"github.com/actiontech/dms/api"
	"github.com/labstack/echo/v4"
)

func (d *DMSService) RegisterSwagger(c echo.Context) error {
	targets, err := d.DmsProxyUsecase.ListProxyTargets(c.Request().Context())
	if err != nil {
		return err
	}

	for _, target := range targets {
		targetName := api.SwaggerType(target.Name)
		_, ok := api.GetSwaggerDoc(targetName)
		if !ok {
			reply, err := d.SwaggerUseCase.GetSwaggerContentByType(c, targetName)
			if err != nil {
				return err
			}
			api.RegisterSwaggerDoc(targetName, reply)
		}
	}

	return nil
}
