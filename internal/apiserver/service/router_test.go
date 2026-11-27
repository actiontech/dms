package service

import (
	"testing"

	"github.com/actiontech/dms/internal/dms/biz"
	dmsService "github.com/actiontech/dms/internal/dms/service"
	sql_workbench "github.com/actiontech/dms/internal/sql_workbench/service"
	"github.com/stretchr/testify/assert"
)

func TestBuildSQLWorkbenchProxyConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		controller *SqlWorkbenchController
		expected   sqlWorkbenchProxyConfiguration
	}{
		"nil controller": {
			controller: nil,
			expected:   sqlWorkbenchProxyConfiguration{},
		},
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
			expected: sqlWorkbenchProxyConfiguration{
				enableCloudbeaver: true,
				cloudbeaverRoot:   "/sql_query",
				enableOdcQuery:    false,
				odcQueryRoot:      "/odc_query",
			},
		},
		"all disabled": {
			controller: &SqlWorkbenchController{
				CloudbeaverService: &dmsService.CloudbeaverService{
					CloudbeaverUsecase: biz.NewCloudbeaverUsecase(noopLogger{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
				},
				SqlWorkbenchService: &sql_workbench.SqlWorkbenchService{},
			},
			expected: sqlWorkbenchProxyConfiguration{
				enableCloudbeaver: false,
				cloudbeaverRoot:   "/sql_query",
				enableOdcQuery:    false,
				odcQueryRoot:      "/odc_query",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, buildSQLWorkbenchProxyConfiguration(tc.controller))
		})
	}
}
