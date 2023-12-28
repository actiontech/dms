package biz

import (
	"context"
	"errors"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type DMSConfig struct {
	Base
	UID                                   string `json:"uid"`
	NeedInitOpPermissions                 bool   `json:"need_init_op_permissions"`
	NeedInitUsers                         bool   `json:"need_init_users"`
	NeedInitRoles                         bool   `json:"need_init_roles"`
	NeedInitProjects                      bool   `json:"need_init_projects"`
	EnableSQLResultSetsDataLossProtection bool   `json:"enable_sql_result_sets_data_loss_protection"`
}

type DMSConfigRepo interface {
	GetDMSConfig(ctx context.Context, uid string) (*DMSConfig, error)
	SaveDMSConfig(ctx context.Context, dmsConfig *DMSConfig) error
	UpdateDMSConfig(ctx context.Context, dmsConfig *DMSConfig) error
}

type DMSConfigUseCase struct {
	repo DMSConfigRepo
	log  *utilLog.Helper
}

func NewDMSConfigUseCase(log utilLog.Logger, repo DMSConfigRepo) *DMSConfigUseCase {
	return &DMSConfigUseCase{repo: repo, log: utilLog.NewHelper(log, utilLog.WithMessageKey("biz.dms_config"))}
}

func (n *DMSConfigUseCase) GetDMSConfig(ctx context.Context) (*DMSConfig, error) {
	conf, err := n.repo.GetDMSConfig(ctx, pkgConst.UIDOfDMSConfig)
	if nil != err {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			// 如果没有找到，则直接初始化
			dms := &DMSConfig{
				UID:                                   pkgConst.UIDOfDMSConfig,
				NeedInitOpPermissions:                 true,
				NeedInitUsers:                         true,
				NeedInitRoles:                         true,
				NeedInitProjects:                      true,
				EnableSQLResultSetsDataLossProtection: false,
			}
			if err := n.SaveDMSConfig(ctx, dms); nil != err {
				return nil, err
			}
			return dms, nil
		}
		return nil, err
	}
	return conf, nil
}

func (n *DMSConfigUseCase) SaveDMSConfig(ctx context.Context, dmsConfig *DMSConfig) error {
	return n.repo.SaveDMSConfig(ctx, dmsConfig)
}

func (n *DMSConfigUseCase) UpdateDMSConfig(ctx context.Context, dmsConfig *DMSConfig) error {
	return n.repo.UpdateDMSConfig(ctx, dmsConfig)
}

func (n *DMSConfigUseCase) IsEnableSQLResultsDataLossProtection(ctx context.Context) (bool, error) {
	conf, err := n.GetDMSConfig(ctx)
	if nil != err {
		return false, err
	}
	return conf.EnableSQLResultSetsDataLossProtection, nil
}
