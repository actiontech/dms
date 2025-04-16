package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type ResourceOverviewRepo interface {
	GetResourceOverviewTopology(ctx context.Context, listOptions ListResourceOverviewOption) (*ResourceTopology, error)
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
	FilterByDBType            string `json:"filter_by_db_type"`
	FilterByBusinessTagUID    string `json:"filter_by_business_tag_uid"`
	FilterByEnvironmentTagUID string `json:"filter_by_environment_tag_uid"`
	FilterByProjectUID        string `json:"filter_by_project_uid"`
	FilterByProjectUIDs        []string `json:"filter_by_project_uids"`
	FuzzySearchResourceName   string `json:"fuzzy_search_resource_name"`
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

type ResourceOverviewVisibility string
