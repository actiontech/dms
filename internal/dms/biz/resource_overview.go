package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type ResourceOverviewVisibility string

type ResourceOverviewRepo interface {
	GetResourceList(ctx context.Context, listOptions ListResourceOverviewOption) ([]*ResourceRow, error)
}

type ResourceTopology struct {
	Resources []*Resource `json:"resources"`
}

type Resource struct {
	Business Business           `json:"business"`
	Projects []*ResourceProject `json:"project"`
}

type ResourceProject struct {
	Project    Project      `json:"project"`
	DBServices []*DBService `json:"db_services"`
}

type ListResourceOverviewOption struct {
	ListOptions *ResourceOverviewListOptions
	Filters     *ResourceOverviewFilter
}

type ResourceOverviewListOptions struct {
	SortByField string `json:"sort_by_field"`
	SortAsc     bool   `json:"sort_asc"`
	PageIndex   uint32 `json:"page_index"`
	PageSize    uint32 `json:"page_size"`
}

type ResourceOverviewFilter struct {
	FilterByDBType            string   `json:"filter_by_db_type"`
	FilterByBusinessTagUID    string   `json:"filter_by_business_tag_uid"`
	FilterByEnvironmentTagUID string   `json:"filter_by_environment_tag_uid"`
	FilterByProjectUID        string   `json:"filter_by_project_uid"`
	FilterByProjectUIDs       []string `json:"filter_by_project_uids"`
	FuzzySearchResourceName   string   `json:"fuzzy_search_resource_name"`
}

type ResourceRow struct {
	ProjectName        string `json:"project_name"`
	ProjectUID         string `json:"project_uid"`
	EnvironmentTagUID  string `json:"environment_tag_uid"`
	EnvironmentTagName string `json:"environment_tag_name"`
	BusinessTagName    string `json:"business_tag_name"`
	BusinessTagUID     string `json:"business_tag_uid"`
	DBServiceName      string `json:"db_service_name"`
	DBServiceUID       string `json:"db_service_uid"`
	DBType             string `json:"db_type"`
}

type ResourceOverviewUsecase struct {
	log                       *utilLog.Helper
	projectRepo               ProjectRepo
	dbServiceRepo             DBServiceRepo
	resourceOverviewRepo      ResourceOverviewRepo
	opPermissionVerifyUsecase OpPermissionVerifyUsecase
}

func NewResourceOverviewUsecase(
	log utilLog.Logger,
	projectRepo ProjectRepo,
	dbServiceRepo DBServiceRepo,
	opPermissionVerifyUsecase OpPermissionVerifyUsecase,
	resourceOverviewRepo ResourceOverviewRepo,
) *ResourceOverviewUsecase {
	return &ResourceOverviewUsecase{
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.ResourceOverview")),
		projectRepo:               projectRepo,
		dbServiceRepo:             dbServiceRepo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		resourceOverviewRepo:      resourceOverviewRepo,
	}
}
