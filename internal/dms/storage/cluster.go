package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.ClusterRepo = (*ClusterRepo)(nil)

type ClusterRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewClusterRepo(log utilLog.Logger, s *Storage) *ClusterRepo {
	return &ClusterRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.cluster"))}
}

func (d *ClusterRepo) GetClusterNodes(ctx context.Context) ([]*biz.ClusterNodeInfo, error) {
	var items []*model.ClusterNodeInfo
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		db := tx.WithContext(ctx)

		if err := db.Find(&items).Error; err != nil {
			return fmt.Errorf("failed to list cluster nodes: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.ClusterNodeInfo, 0, len(items))
	// convert model to biz
	for _, item := range items {
		ds, err := convertModelClusterNodeInfo(item)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert cluster_node_info: %w", err))
		}
		ret = append(ret, ds)
	}

	return ret, nil
}

func (d *ClusterRepo) RegisterClusterNode(ctx context.Context, params *biz.ClusterNodeInfo) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizCLusterNodeInfo(params)).Error; err != nil {
			return fmt.Errorf("failed to save cluster_node_info: %v", err)
		}

		return nil
	})
}

func (d *ClusterRepo) GetClusterLeader(ctx context.Context) (*biz.ClusterLeader, error) {
	var item *model.ClusterLeader
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		if err := tx.WithContext(ctx).Find(&item).Error; err != nil {
			return fmt.Errorf("failed to get cluster leader: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return convertModelClusterLeader(item)
}

const leaderTableAnchor = 1

func (d *ClusterRepo) MaintainClusterLeader(ctx context.Context, serverId string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		var maintainClusterLeaderSql = `
INSERT ignore INTO cluster_leaders (anchor, server_id, last_seen_time) VALUES (?, ?, now()) 
ON DUPLICATE KEY UPDATE 
server_id = IF(last_seen_time < now() - interval 30 second, VALUES(server_id), server_id), 
last_seen_time = IF(server_id = VALUES(server_id), VALUES(last_seen_time), last_seen_time)
`
		return tx.WithContext(ctx).Exec(maintainClusterLeaderSql, leaderTableAnchor, serverId).Error
	})
}
