package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/internal/dms/biz"
	dmsService "github.com/actiontech/dms/internal/dms/service"
	sql_workbench "github.com/actiontech/dms/internal/sql_workbench/service"
	bV1 "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type noopLogger struct{}

func (noopLogger) Log(level utilLog.Level, keyvals ...interface{}) error {
	return nil
}

func TestSqlWorkbenchControllerBuildSQLQueryConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		controller *SqlWorkbenchController
		expected   SQLQueryConfiguration
	}{
		"cloudbeaver enabled": {
			controller: &SqlWorkbenchController{
				CloudbeaverService: &dmsService.CloudbeaverService{
					CloudbeaverUsecase: biz.NewCloudbeaverUsecase(noopLogger{}, &biz.CloudbeaverCfg{
						Host:          "cloudbeaver",
						Port:          "8978",
						AdminUser:     "admin",
						AdminPassword: "password",
					}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
				},
				SqlWorkbenchService: &sql_workbench.SqlWorkbenchService{},
			},
			expected: SQLQueryConfiguration{
				EnableSQLQuery:  true,
				SQLQueryRootURI: "/sql_query/",
				EnableOdcQuery:  false,
				OdcQueryRootURI: "/odc_query",
			},
		},
		"all disabled": {
			controller: &SqlWorkbenchController{
				CloudbeaverService: &dmsService.CloudbeaverService{
					CloudbeaverUsecase: biz.NewCloudbeaverUsecase(noopLogger{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
				},
				SqlWorkbenchService: &sql_workbench.SqlWorkbenchService{},
			},
			expected: SQLQueryConfiguration{
				EnableSQLQuery:  false,
				SQLQueryRootURI: "/sql_query/",
				EnableOdcQuery:  false,
				OdcQueryRootURI: "/odc_query",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.controller.buildSQLQueryConfiguration())
		})
	}
}

func TestGetSQLQueryConfiguration(t *testing.T) {
	t.Parallel()

	controller := &SqlWorkbenchController{
		CloudbeaverService: &dmsService.CloudbeaverService{
			CloudbeaverUsecase: biz.NewCloudbeaverUsecase(noopLogger{}, &biz.CloudbeaverCfg{
				Host:          "cloudbeaver",
				Port:          "8978",
				AdminUser:     "admin",
				AdminPassword: "password",
			}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
		},
		SqlWorkbenchService: &sql_workbench.SqlWorkbenchService{},
	}

	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/dms/configurations/sql_query", nil)
	c := e.NewContext(req, rec)

	err := controller.GetSQLQueryConfiguration(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		bV1.GenericResp
		Data SQLQueryConfiguration `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int(apiError.StatusOK), resp.Code)
	assert.Equal(t, SQLQueryConfiguration{
		EnableSQLQuery:  true,
		SQLQueryRootURI: "/sql_query/",
		EnableOdcQuery:  false,
		OdcQueryRootURI: "/odc_query",
	}, resp.Data)
}
