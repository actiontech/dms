//go:build !enterprise

package biz

import (
	"bytes"
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/labstack/echo/v4"
)

func (cu *CloudbeaverUsecase) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService, userId string) ([]*DBService, error) {
	return activeDBServices, nil
}

func (cu *CloudbeaverUsecase) UpdateCbOp(ctx context.Context, resp cloudbeaver.AuditResults, taskID *string) {
	return
}

func (cu *CloudbeaverUsecase) UpdateCbOpResult(c echo.Context, cloudbeaverResBuf *bytes.Buffer, params *graphql.RawParams, ctx context.Context) error {
	return nil
}

func (cu *CloudbeaverUsecase) SaveCbOpLog(c echo.Context, dbService *DBService, params *graphql.RawParams, next echo.HandlerFunc) (*string, error) {
	return nil, nil
}

func (cu *CloudbeaverUsecase) SaveUiOp(c echo.Context, buf *bytes.Buffer, params *graphql.RawParams) error {
	return nil
}
