package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListCBOperationLogs
type ListCBOperationLogsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in:query
	FilterOperationPersonUID string `json:"filter_operation_person_uid" query:"filter_operation_person_uid"`
	// in:query
	FilterOperationTimeFrom string `json:"filter_operation_time_from" query:"filter_operation_time_from"`
	// in:query
	FilterOperationTimeTo string `json:"filter_operation_time_to" query:"filter_operation_time_to"`
	// in:query
	FilterDBServiceUID string `json:"filter_db_service_uid" query:"filter_db_service_uid"`
	// in:query
	FilterExecResult string `json:"filter_exec_result" query:"filter_exec_result"`
	// filter fuzzy key word for operation_detail/operation_ip
	// in:query
	FuzzyKeyword string `json:"fuzzy_keyword" query:"fuzzy_keyword"`
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:model ListCBOperationLogsReply
type ListCBOperationLogsReply struct {
	// 执行SQL总量
	ExecSQLTotal int64 `json:"exec_sql_total"`
	// 执行成功率
	ExecSuccessRate float64 `json:"exec_success_rate"`
	// 审核拦截的异常SQL数量
	AuditInterceptedSQLCount int64 `json:"audit_intercepted_sql_count"`
	// 执行失败的SQL
	ExecFailedSQLCount int64 `json:"exec_failed_sql_count"`
	// list cb operation logs reply
	Data  []*CBOperationLog `json:"data"`
	Total int64             `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

type CBOperationLog struct {
	UID               string               `json:"uid"`
	OperationPerson   UidWithName          `json:"operation_person"`
	OperationTime     time.Time            `json:"operation_time"`
	DBService         UidWithDBServiceName `json:"db_service"`
	OperationDetail   string               `json:"operation_detail"`
	SessionID         string               `json:"session_id"`
	OperationIp       string               `json:"operation_ip"`
	AuditResult       []*AuditSQLResult    `json:"audit_result"`
	ExecResult        string               `json:"exec_result"`
	ExecTimeSecond    int                  `json:"exec_time_second"`
	ResultSetRowCount int64                `json:"result_set_row_count"`
}

type UidWithDBServiceName struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}
