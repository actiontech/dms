package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:enum OpRangeType
type OpRangeType string

const (
	OpRangeTypeUnknown OpRangeType = "unknown"
	// 全局权限: 该权限只能被用户使用
	OpRangeTypeGlobal OpRangeType = "global"
	// 项目权限: 该权限只能被成员使用
	OpRangeTypeProject OpRangeType = "project"
	// 项目内的数据源权限: 该权限只能被成员使用
	OpRangeTypeDBService OpRangeType = "db_service"
)

func ParseOpRangeType(typ string) (OpRangeType, error) {
	switch typ {
	case string(OpRangeTypeDBService):
		return OpRangeTypeDBService, nil
	case string(OpRangeTypeProject):
		return OpRangeTypeProject, nil
	case string(OpRangeTypeGlobal):
		return OpRangeTypeGlobal, nil
	default:
		return "", fmt.Errorf("invalid op range type: %s", typ)
	}
}

// swagger:parameters ListOpPermissions
type ListOpPermissionReq struct {
	// the maximum count of op permission to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of op permissions to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Order by the specified field
	// in:query
	OrderBy OpPermissionOrderByField `query:"order_by" json:"order_by"`
	// filter by op permission target
	// in:query
	FilterByTarget OpPermissionTarget `query:"filter_by_target" json:"filter_by_target"  validate:"required"`
}

// swagger:enum OpPermissionTarget
type OpPermissionTarget string

const (
	OpPermissionTargetAll    OpPermissionTarget = "all"
	OpPermissionTargetUser   OpPermissionTarget = "user"
	OpPermissionTargetMember OpPermissionTarget = "member"
	OpPermissionTargetProject OpPermissionTarget = "project"
)

// swagger:enum OpPermissionOrderByField
type OpPermissionOrderByField string

const (
	OpPermissionOrderByName OpPermissionOrderByField = "name"
)

// A dms op permission
type ListOpPermission struct {
	// op permission
	OpPermission UidWithName `json:"op_permission"`
	Module       string		 `json:"module"`
	Description  string      `json:"description"`
	RangeType    OpRangeType `json:"range_type"`
}

// swagger:model ListOpPermissionReply
type ListOpPermissionReply struct {
	// List op_permission reply
	Data  []*ListOpPermission `json:"data"`
	Total int64               `json:"total_nums"`

	// Generic reply
	base.GenericResp
}
