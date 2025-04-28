package biz

import (
	"context"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

type GatewayRepo interface {
	AddGateway(ctx context.Context, u *Gateway) error
	DeleteGateway(ctx context.Context, id string) error
	UpdateGateway(ctx context.Context, u *Gateway) error
	GetGateway(ctx context.Context, id string) (*Gateway, error)
	ListGateways(ctx context.Context, opt *ListGatewaysOption) ([]*Gateway, int64, error)
	GetGatewayTips(ctx context.Context) ([]*Gateway, error)
	SyncGateways(ctx context.Context, s []*Gateway) error
}

type Gateway struct {
	ID   string
	Name string
	Desc string
	URL  string
}

type ListGatewaysOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      MemberField
	FilterBy     []pkgConst.FilterCondition
}
