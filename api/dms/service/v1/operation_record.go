package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
)

// swagger:model
type AddOperationRecordReq struct {
	OperationRecord *OperationRecord `json:"operation_record" validate:"required"`
}

// swagger:model AddOperationRecordReply
type AddOperationRecordReply struct {
	base.GenericResp
}

type OperationRecord struct {
	OperationTime        time.Time       `json:"operation_time"`
	OperationUserName    string          `json:"operation_user_name" validate:"required"`
	OperationReqIP       string          `json:"operation_req_ip"`
	OperationUserAgent   string          `json:"operation_user_agent"`
	OperationTypeName    string          `json:"operation_type_name"`
	OperationAction      string          `json:"operation_action"`
	OperationProjectName string          `json:"operation_project_name"`
	OperationStatus      string          `json:"operation_status"`
	OperationI18nContent i18nPkg.I18nStr `json:"operation_i18n_content"`
}

// swagger:parameters GetOperationRecordList
type GetOperationRecordListReq struct {
	// in:query
	FilterOperateTimeFrom string `json:"filter_operate_time_from" query:"filter_operate_time_from"`
	// in:query
	FilterOperateTimeTo string `json:"filter_operate_time_to" query:"filter_operate_time_to"`
	// in:query
	FilterOperateProjectName *string `json:"filter_operate_project_name" query:"filter_operate_project_name"`
	// in:query
	FuzzySearchOperateUserName string `json:"fuzzy_search_operate_user_name" query:"fuzzy_search_operate_user_name"`
	// in:query
	FuzzySearchOperateContent string `json:"fuzzy_search_operate_content" query:"fuzzy_search_operate_content"`
	// in:query
	FilterFuzzyOperateUserName string `json:"filter_fuzzy_operate_user_name" query:"filter_fuzzy_operate_user_name"`
	// in:query
	FilterOperateTypeNames []string `json:"filter_operate_type_names" query:"filter_operate_type_names"`
	// in:query
	FilterOperateActions []string `json:"filter_operate_actions" query:"filter_operate_actions"`
	// in:query
	FilterOperateStatus string `json:"filter_operate_status" query:"filter_operate_status"`
	// in:query
	// Required: true
	PageIndex uint32 `json:"page_index" query:"page_index" validate:"required"`
	// in:query
	// Required: true
	PageSize uint32 `json:"page_size" query:"page_size" validate:"required"`
}

// swagger:model GetOperationRecordListReply
type GetOperationRecordListReply struct {
	Data      []OperationRecordListItem `json:"data"`
	TotalNums uint64                    `json:"total_nums"`
	base.GenericResp
}

type OperationRecordListItem struct {
	ID                 uint64        `json:"id"`
	OperationTime      *time.Time    `json:"operation_time"`
	OperationUser      OperationUser `json:"operation_user"`
	OperationUserAgent string        `json:"operation_user_agent"`
	OperationTypeName  string        `json:"operation_type_name"`
	OperationAction    string        `json:"operation_action"`
	OperationContent   string        `json:"operation_content"`
	ProjectName        string        `json:"project_name"`
	// enum: succeeded,failed
	Status string `json:"status"`
}

type OperationUser struct {
	UserName string `json:"user_name"`
	IP       string `json:"ip"`
}

// swagger:model GetOperationTypeNameListReply
type GetOperationTypeNameListReply struct {
	Data []OperationTypeNameListItem `json:"data"`
	base.GenericResp
}

type OperationTypeNameListItem struct {
	OperationTypeName string `json:"operation_type_name"`
	Desc              string `json:"desc"`
}

// swagger:model GetOperationActionListReply
type GetOperationActionListReply struct {
	Data []OperationActionListItem `json:"data"`
	base.GenericResp
}

type OperationActionListItem struct {
	OperationType   string `json:"operation_type"`
	OperationAction string `json:"operation_action"`
	Desc            string `json:"desc"`
}

// swagger:model GetOperationUserNameListReply
type GetOperationUserNameListReply struct {
	Data []OperationUserNameListItem `json:"data"`
	base.GenericResp
}

type OperationUserNameListItem struct {
	OperationUserName string `json:"operation_user_name"`
	OperationReqIP    string `json:"operation_req_ip"`
	Desc              string `json:"desc"`
}

// swagger:parameters GetOperationUserNameList
type GetOperationUserNameListReq struct {
	// in:query
	FilterOperateProjectName *string `json:"filter_operate_project_name" query:"filter_operate_project_name"`
}

// swagger:parameters ExportOperationRecordList
type ExportOperationRecordListReq struct {
	// in:query
	FilterOperateTimeFrom string `json:"filter_operate_time_from" query:"filter_operate_time_from"`
	// in:query
	FilterOperateTimeTo string `json:"filter_operate_time_to" query:"filter_operate_time_to"`
	// in:query
	FilterOperateProjectName *string `json:"filter_operate_project_name" query:"filter_operate_project_name"`
	// in:query
	FuzzySearchOperateUserName string `json:"fuzzy_search_operate_user_name" query:"fuzzy_search_operate_user_name"`
	// in:query
	FuzzySearchOperateContent string `json:"fuzzy_search_operate_content" query:"fuzzy_search_operate_content"`
	// in:query
	FilterFuzzyOperateUserName string `json:"filter_fuzzy_operate_user_name" query:"filter_fuzzy_operate_user_name"`
	// in:query
	FilterOperateTypeNames []string `json:"filter_operate_type_names" query:"filter_operate_type_names"`
	// in:query
	FilterOperateActions []string `json:"filter_operate_actions" query:"filter_operate_actions"`
	// in:query
	FilterOperateStatus string `json:"filter_operate_status" query:"filter_operate_status"`
}

// swagger:response ExportOperationRecordListReply
type ExportOperationRecordListReply struct {
	// swagger:file
	// in: body
	File []byte
}
