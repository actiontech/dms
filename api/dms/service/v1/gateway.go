package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:model
type AddGatewayReq struct {
	AddGateway *Gateway `json:"add_gateway"  validate:"required"`
}

// swagger:model AddGatewayReply
type AddGatewayReply struct {
	base.GenericResp
}

// swagger:parameters DeleteGateway
type DeleteGatewayReq struct {
	// in:path
	// Required: true
	DeleteGatewayID string `param:"gateway_id" json:"gateway_id"  validate:"required"`
}

// swagger:model DeleteGatewayReply
type DeleteGatewayReply struct {
	base.GenericResp
}

// swagger:model
type UpdateGatewayReq struct {
	// swagger:ignore
	UpdateGatewayID string        `param:"gateway_id" json:"gateway_id"  validate:"required"`
	UpdateGateway   UpdateGateway `json:"update_gateway"  validate:"required"`
}

type UpdateGateway struct {
	GatewayName    string `json:"gateway_name"  validate:"required"`
	GatewayDesc    string `json:"gateway_desc"`
	GatewayAddress string `json:"gateway_address" binding:"required"`
}

// swagger:model UpdateGatewayReply
type UpdateGatewayReply struct {
	base.GenericResp
}

// swagger:parameters GetGateway
type GetGatewayReq struct {
	// in:path
	// Required: true
	GetGatewayID string `param:"gateway_id" json:"gateway_id"  validate:"required"`
}

type Gateway struct {
	GatewayID      string `json:"gateway_id"  validate:"required"`
	GatewayName    string `json:"gateway_name"  validate:"required"`
	GatewayDesc    string `json:"gateway_desc" `
	GatewayAddress string `json:"gateway_address" binding:"required"`
}

// swagger:model GetGatewayReply
type GetGatewayReply struct {
	Gateways *Gateway `json:"data"`
	base.GenericResp
}

// swagger:parameters ListGateways
type ListGatewaysReq struct {
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
}

// swagger:model ListGatewaysReply
type ListGatewaysReply struct {
	Total    int64      `json:"total"`
	Gateways []*Gateway `json:"data"`
	base.GenericResp
}

// swagger:model GetGatewayTipsReply
type GetGatewayTipsReply struct {
	base.GenericResp
	GatewayTips []*UidWithName `json:"data"`
}

// swagger:model
type SyncGatewayReq struct {
	Gateways []*Gateway `json:"gateways"  validate:"required"`
}
