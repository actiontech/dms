package rand

import (
	"github.com/bwmarrin/snowflake"
)

var defaultNodeNo int64 = 1
var node *snowflake.Node

// InitSnowflake initiate Snowflake node singleton.
func InitSnowflake(nodeNo int64) error {
	// Create snowflake node
	n, err := snowflake.NewNode(nodeNo)
	if err != nil {
		return err
	}
	// Set node
	node = n
	return nil
}

// genUid为生成随机uid
func GenUid() (int64, error) {
	if node == nil {
		if err := InitSnowflake(defaultNodeNo); err != nil {
			return 0, err
		}
	}
	return node.Generate().Int64(), nil
}

// genStrUid为生成随机uid的字符串形式
func GenStrUid() (string, error) {
	if node == nil {
		if err := InitSnowflake(defaultNodeNo); err != nil {
			return "", err
		}
	}
	return node.Generate().String(), nil
}
