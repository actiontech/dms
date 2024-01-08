package biz

import (
	"context"
	"time"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type ClusterRepo interface {
	GetClusterLeader(ctx context.Context) (*ClusterLeader, error)
	MaintainClusterLeader(ctx context.Context, serverId string) error
	GetClusterNodes(ctx context.Context) ([]*ClusterNodeInfo, error)
	RegisterClusterNode(ctx context.Context, params *ClusterNodeInfo) error
}

type ClusterUsecase struct {
	tx       TransactionGenerator
	repo     ClusterRepo
	log      *utilLog.Helper
	serverId string
	exitCh   chan struct{}
	doneCh   chan struct{}
}

func NewClusterUsecase(log utilLog.Logger, tx TransactionGenerator, repo ClusterRepo) *ClusterUsecase {
	return &ClusterUsecase{
		tx:     tx,
		repo:   repo,
		log:    utilLog.NewHelper(log, utilLog.WithMessageKey("biz.cluster")),
		exitCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

type ClusterLeader struct {
	Anchor       int       `json:"anchor"`
	ServerId     string    `json:"server_id"`
	LastSeenTime time.Time `json:"last_seen_time"`
}

type ClusterNodeInfo struct {
	ServerId     string `json:"server_id"`
	HardwareSign string `json:"hardware_sign"`
}

var clusterMode bool = false

type ClusterImpl interface {
	Join(serverId string) error
	Leave()
	IsLeader() bool
	IsClusterMode() bool
	SetClusterMode(mode bool)
}
