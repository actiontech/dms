package biz

import (
	"github.com/actiontech/dms/api"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

type SwaggerUseCase struct {
	log             *utilLog.Helper
	dmsProxyUsecase *DmsProxyUsecase
}

func NewSwaggerUseCase(logger utilLog.Logger, dmsProxyUseCase *DmsProxyUsecase) *SwaggerUseCase {
	return &SwaggerUseCase{
		log:             utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.swagger")),
		dmsProxyUsecase: dmsProxyUseCase,
	}
}

func (s *SwaggerUseCase) GetSwaggerContentByType(c echo.Context, targetName api.SwaggerType) ([]byte, error) {
	target, err := s.dmsProxyUsecase.GetTargetByName(c.Request().Context(), string(targetName))
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
