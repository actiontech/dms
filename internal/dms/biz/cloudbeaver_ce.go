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

func (cu *CloudbeaverUsecase) UpdateCbOpResult(c echo.Context, cloudbeaverResBuf *bytes.Buffer, params *graphql.RawParams, ctx context.Context) error {
	return nil
}

func (cu *CloudbeaverUsecase) SaveCbOpLog(c echo.Context, dbService *DBService, params *graphql.RawParams, auditResult []cloudbeaver.AuditSQLResV2, isAuditPass bool, taskID *string) error {
	return nil
}

func (cu *CloudbeaverUsecase) SaveCbOpLogForWorkflow(c echo.Context, dbService *DBService, params *graphql.RawParams, auditResult []cloudbeaver.AuditSQLResV2, isAuditPass bool, workflowID string, isExecFailed bool) error {
	return nil
}

func (cu *CloudbeaverUsecase) SaveUiOp(c echo.Context, buf *bytes.Buffer, params *graphql.RawParams) error {
	return nil
}

func (cu *CloudbeaverUsecase) SaveCbLogSqlAuditNotEnable(c echo.Context, dbService *DBService, params *graphql.RawParams, cloudbeaverResBuf *bytes.Buffer) error {
	return nil
}
