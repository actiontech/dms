package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:parameters GetCBDbServiceStatistic
type GetCBDbServiceStatisticReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetCBDbServiceStatisticReply
type GetCBDbServiceStatisticReply struct {
	// Generic reply
	base.GenericResp
	Data []*CbDbServiceStatistic `json:"data"`
}

type CbDbServiceStatistic struct {
	Name    string                         `json:"name"`
	Count   int64                          `json:"count"`
	Content []*CbDbServiceStatisticContent `json:"content"`
}

type CbDbServiceStatisticContent struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
	Count  int64  `json:"count"`
}

// swagger:parameters GetCBOperationStatistic
type GetCBOperationStatisticReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetCBOperationStatisticReply
type GetCBOperationStatisticReply struct {
	// Generic reply
	base.GenericResp
	Data []*CbOperationStatistic `json:"data"`
}

type CbOperationStatistic struct {
	OperationType  string `json:"operation_type"`
	OperationCount int64  `json:"operation_count"`
}
