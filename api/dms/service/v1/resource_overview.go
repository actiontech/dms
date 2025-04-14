package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// 资源概览接口组 API Model

// 资源概览统计接口
// route: /v1/dms/resource_overview/statistics
// Method: GET

// swagger:parameters GetResourceOverviewStatisticsV1
type ResourceOverviewStatisticsReq struct{}

// swagger:model ResourceOverviewStatisticsResV1
type ResourceOverviewStatisticsRes struct {
	Data struct {
		// 业务总数
		BusinessTotalNumber int64 `json:"business_total_number"`
		// 项目总数
		ProjectTotalNumber int64 `json:"project_total_number"`
		// 数据源总数
		DBServiceTotalNumber int64 `json:"db_service_total_number"`
	} `json:"data"`
	base.GenericResp
}

// 资源类型分布接口
// route: /v1/dms/resource_overview/resource_type_distribution
// Method: GET

// swagger:parameters GetResourceOverviewResourceTypeDistributionV1
type ResourceOverviewResourceTypeDistributionReq struct{}

// swagger:model ResourceOverviewResourceTypeDistributionResV1
type ResourceOverviewResourceTypeDistributionRes struct {
	Data  []*ResourceTypeDistributionData `json:"data"`
	Total int64                           `json:"total_nums"`
	base.GenericResp
}

// swagger:model ResourceTypeDistributionData
type ResourceTypeDistributionData struct {
	// 资源类型
	ResourceType string `json:"resource_type"`
	// 数量
	Count int64 `json:"count"`
}

// 资源概览拓扑接口
// route: /v1/dms/resource_overview/topology
// Method: GET

// swagger:parameters GetResourceOverviewTopologyV1
type ResourceOverviewTopologyReq struct {
	ResourceOverviewResourceListFilter
}

// swagger:parameters
type ResourceOverviewResourceListFilter struct {
	// 根据数据源类型筛选
	// in:query
	// type:string
	DBType string `param:"db_type" json:"db_type"`
	// 根据所属业务标签筛选
	// in:query
	// type:string
	BusinessTagUID string `param:"business_tag_uid" json:"business_tag_uid"`
	// 根据环境属性标签筛选
	// in:query
	// type:string
	EnvironmentTagUID string `param:"environment_tag_uid" json:"environment_tag_uid"`
	// 根据所属项目筛选
	// in:query
	// type:string
	ProjectUID string `param:"project_uid" json:"project_uid"`
	// 根据某列排序 enums:"audit_score,pending_workflow_count,high_priority_sql_count"
	// in:query
	// type:string
	SortByField string `query:"sort_by_field" json:"sort_by_field" enums:"audit_score,pending_workflow_count,high_priority_sql_count"`
	// 是否正序排序
	// in:query
	// type:bool
	SortAsc bool `query:"sort_asc" json:"sort_asc"`
}

// swagger:enum ResourceListSortByField
type ResourceListSortByField string

const (
	// 根据审核评分排序
	SortByFieldAuditScore ResourceListSortByField = "audit_score"
	// 根据待处理工单数排序
	SortByFieldPendingWorkflowCount ResourceListSortByField = "pending_workflow_count"
	// 根据高优先级SQL数排序
	SortByFieldHighPrioritySQLCount ResourceListSortByField = "high_priority_sql_count"
)

// swagger:model ResourceOverviewTopologyResV1
type ResourceOverviewTopologyRes struct {
	// business:project = 1:n project:db_service = 1:n
	Data  []*Business `json:"data"`
	Total int64       `json:"total_nums"`
	base.GenericResp
}

// swagger:model ResourceBusiness
type Business struct {
	BusinessTag *BusinessTag       `json:"business_tag"`
	Project     []*ResourceProject `json:"project"`
}

// swagger:model ResourceProject
type ResourceProject struct {
	ProjectUID  string               `json:"project_uid"`
	ProjectName string               `json:"project_name"`
	DBService   []*ResourceDBService `json:"db_service"`
}

// swagger:model ResourceDBService
type ResourceDBService struct {
	DBServiceUID  string `json:"db_service_uid"`
	DBServiceName string `json:"db_service_name"`
}

// 资源详情列表接口
// route: /v1/dms/resource_overview/resource_list
// Method: GET

// swagger:parameters GetResourceOverviewResourceListV1
type ResourceOverviewResourceListReq struct {
	ResourceOverviewResourceListFilter
}

// swagger:model ResourceOverviewResourceListResV1
type ResourceOverviewResourceListRes struct {
	Data  []*ResourceListData `json:"data"`
	Total int64               `json:"total_nums"`
	base.GenericResp
}

// swagger:model ResourceListData
type ResourceListData struct {
	// 资源类型
	ResourceType string `json:"resource_type"`
	// 资源名称
	ResourceName string `json:"resource_name"`
	// 所属业务
	BusinessTag *BusinessTag `json:"business_tag"`
	// 所属项目
	Project *ResourceProject `json:"project"`
	// 审核评分
	AuditScore int64 `json:"audit_score"`
	// 待处理工单数
	PendingWorkflowCount int64 `json:"pending_workflow_count"`
	// 高优先级SQL数
	HighPrioritySQLCount int64 `json:"high_priority_sql_count"`
}
