package service

import (
	"github.com/actiontech/dms/api"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/labstack/echo/v4"
)

func (d *DMSService) RegisterSwagger(c echo.Context) error {
	targets, err := d.DmsProxyUsecase.ListProxyTargetsByScenarios(c.Request().Context(), []biz.ProxyScenario{biz.ProxyScenarioInternalService})
	if err != nil {
		return err
	}

	for _, target := range targets {
		targetName := api.SwaggerType(target.Name)
		_, ok := api.GetSwaggerDoc(targetName)
		if !ok {
			reply, err := d.SwaggerUseCase.GetSwaggerContentByType(c, targetName)
			if err != nil {
				d.log.Errorf("failed to get swagger content by type: %s, err: %v", targetName, err)
				continue
			}
			api.RegisterSwaggerDoc(targetName, reply)
		}
	}

	return nil
}
