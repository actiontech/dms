//go:build !enterprise
// +build !enterprise

package biz

func (c *ClusterUsecase) Join(serverId string) error {
	return nil
}

func (c *ClusterUsecase) Leave() {
}

func (c *ClusterUsecase) IsLeader() bool {
	return false
}

func (c *ClusterUsecase) IsClusterMode() bool {
	return false
}

func (c *ClusterUsecase) SetClusterMode(mode bool) {
}
