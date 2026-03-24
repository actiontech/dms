//go:build !dms

package service

import (
	"context"
	"errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotSupportDataMasking = errors.New("DataMasking related functions are dms version functions")

func (d *DMSService) ConfigureMaskingRules(ctx context.Context, req *v1.ConfigureMaskingRulesReq) error {
	return errNotSupportDataMasking
}

func (d *DMSService) AddSensitiveDataDiscoveryTask(ctx context.Context, req *v1.AddSensitiveDataDiscoveryTaskReq) (reply *v1.AddSensitiveDataDiscoveryTaskReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) UpdateSensitiveDataDiscoveryTask(ctx context.Context, req *v1.UpdateSensitiveDataDiscoveryTaskReq) (reply *v1.UpdateSensitiveDataDiscoveryTaskReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) DeleteSensitiveDataDiscoveryTask(ctx context.Context, req *v1.DeleteSensitiveDataDiscoveryTaskReq) error {
	return errNotSupportDataMasking
}

func (d *DMSService) GetMaskingOverviewTree(ctx context.Context, req *v1.GetMaskingOverviewTreeReq, currentUserUid string) (reply *v1.GetMaskingOverviewTreeReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) GetTableColumnMaskingDetails(ctx context.Context, req *v1.GetTableColumnMaskingDetailsReq) (reply *v1.GetTableColumnMaskingDetailsReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) ListSensitiveDataDiscoveryTasks(ctx context.Context, req *v1.ListSensitiveDataDiscoveryTasksReq) (reply *v1.ListSensitiveDataDiscoveryTasksReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) ListSensitiveDataDiscoveryTaskHistories(ctx context.Context, req *v1.ListSensitiveDataDiscoveryTaskHistoriesReq) (reply *v1.ListSensitiveDataDiscoveryTaskHistoriesReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) ListMaskingRules(ctx context.Context) (reply *dmsV1.ListMaskingRulesReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) ListMaskingTemplates(ctx context.Context, req *dmsV1.ListMaskingTemplatesReq) (reply *dmsV1.ListMaskingTemplatesReply, err error) {
	return nil, errNotSupportDataMasking
}

func (d *DMSService) AddMaskingTemplate(ctx context.Context, req *dmsV1.AddMaskingTemplateReq) error {
	return errNotSupportDataMasking
}

func (d *DMSService) UpdateMaskingTemplate(ctx context.Context, req *dmsV1.UpdateMaskingTemplateReq) error {
	return errNotSupportDataMasking
}

func (d *DMSService) DeleteMaskingTemplate(ctx context.Context, req *dmsV1.DeleteMaskingTemplateReq) error {
	return errNotSupportDataMasking
}

func initDataMaskingUsecase(_ utilLog.Logger, _ *storage.Storage, _ *biz.DBServiceUsecase, _ *biz.ClusterUsecase, _ biz.ProxyTargetRepo) (*dataMaskingUsecase, func(), error) {
	return nil, func() {}, nil
}

func newCloudbeaverSQLResultMasker(_ utilLog.Logger, _ *storage.Storage, _ biz.ProxyTargetRepo) (biz.SQLResultMasker, error) {
	return nil, nil
}

type dataMaskingDiscoveryTaskUsecase interface {
	ListMaskingTaskStatus(ctx context.Context, dbServiceUIDs []string) (map[string]bool, error)
}

type dataMaskingUsecase struct {
	DiscoveryTaskUsecase dataMaskingDiscoveryTaskUsecase
}

func initDataExportMaskingConfigRepo(_ utilLog.Logger, _ *storage.Storage) biz.DataExportMaskingConfigRepo {
	return nil
}
