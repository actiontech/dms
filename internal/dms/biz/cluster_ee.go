//go:build enterprise
// +build enterprise

package biz

import (
	"context"
	"time"

	"github.com/actiontech/dms/internal/license"
)

func (c *ClusterUsecase) IsLeader() bool {
	clusterLeader, err := c.repo.GetClusterLeader(context.TODO())
	if err != nil {
		return false
	}

	return c.serverId == clusterLeader.ServerId
}

func (c *ClusterUsecase) Join(serverId string) error {
	c.serverId = serverId

	hardwareSign, err := license.CollectHardwareInfo()
	if err != nil {
		c.log.Errorf("collect hardware info failed, error: %v", err)
		return err
	}

	params := &ClusterNodeInfo{
		ServerId:     serverId,
		HardwareSign: hardwareSign,
	}
	err = c.repo.RegisterClusterNode(context.TODO(), params)
	if err != nil {
		c.log.Errorf("register cluster node info failed, error: %v", err)
		return err
	}

	err = c.repo.MaintainClusterLeader(context.TODO(), serverId)
	if err != nil {
		c.log.Errorf("maintain cluster leader failed, error: %v", err)
		return err
	}

	go func() {
		tick := time.NewTicker(time.Second * 5)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				err = c.repo.MaintainClusterLeader(context.TODO(), serverId)
				if err != nil {
					c.log.Errorf("maintain cluster leader failed, error: %v", err)
				}
			case <-c.exitCh:
				c.doneCh <- struct{}{}
				return
			}
		}
	}()

	return nil
}

func (c *ClusterUsecase) Leave() {
	c.exitCh <- struct{}{}
	<-c.doneCh
}

func (c *ClusterUsecase) IsClusterMode() bool {
	return clusterMode
}

func (c *ClusterUsecase) SetClusterMode(mode bool) {
	clusterMode = mode
}
