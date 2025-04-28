//go:build !enterprise

package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

func NewDmsGatewayUsecase(logger utilLog.Logger, repo GatewayRepo) (*GatewayUsecase, error) {
	log := utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.gateway"))
	gatewayUsecase := &GatewayUsecase{
		repo:   repo,
		logger: log,
	}
	return gatewayUsecase, nil
}

type GatewayUsecase struct {
	repo   GatewayRepo
	logger *utilLog.Helper
}

func (d *GatewayUsecase) BroadcastAddUser(ctx context.Context, args *CreateUserArgs) error {
	return nil
}

func (d *GatewayUsecase) BroadcastUpdateUser(ctx context.Context, user *User) error {
	return nil
}

func (d *GatewayUsecase) Broadcast() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
