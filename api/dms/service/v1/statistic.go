package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:parameters GetCBInstanceStatistic
type GetCBInstanceStatisticReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetCBInstanceStatisticReply
type GetCBInstanceStatisticReply struct {
	// Generic reply
	base.GenericResp
	Data []*cbInstanceStatistic `json:"data"`
}

type cbInstanceStatistic struct {
	ID      string                        `json:"id"`
	Name    string                        `json:"name"`
	Content []*cbInstanceStatisticContent `json:"content"`
}

type cbInstanceStatisticContent struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
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
	Data []*cbOperationStatistic `json:"data"`
}

type cbOperationStatistic struct {
	OperationType  string `json:"operation_type"`
	OperationCount int64  `json:"operation_count"`
}
