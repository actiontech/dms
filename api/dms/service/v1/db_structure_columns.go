package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListTableColumns
type ListTableColumnsReq struct {
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DBServiceUid string `param:"db_service_uid" json:"db_service_uid" validate:"required"`
	// Required: true
	// in:path
	SchemaName string `param:"schema_name" json:"schema_name" validate:"required"`
	// Required: true
	// in:path
	TableName string `param:"table_name" json:"table_name" validate:"required"`
}

// swagger:model ListTableColumnsReply
type ListTableColumnsReply struct {
	Data []*TableColumn `json:"data"`
	// Generic reply
	base.GenericResp
}

// swagger:model TableColumn
type TableColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Comment  string `json:"comment"`
	Nullable bool   `json:"nullable"`
}
