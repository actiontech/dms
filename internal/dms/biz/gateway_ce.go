//go:build !enterprise

package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func (d *GatewayUsecase) Skipper(c echo.Context) bool {
	return true
}

// AddTarget 实现echo的ProxyBalancer接口，没有实际意义
func (d *GatewayUsecase) Next(c echo.Context) *middleware.ProxyTarget {
	return nil
}

// AddTarget 实现echo的ProxyBalancer接口，没有实际意义
func (d *GatewayUsecase) AddTarget(*middleware.ProxyTarget) bool {
	return true
}

// RemoveTarget 实现echo的ProxyBalancer接口，没有实际意义
func (d *GatewayUsecase) RemoveTarget(string) bool {
	return true
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
