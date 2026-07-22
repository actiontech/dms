package biz

import (
	"fmt"
	"strings"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgParams "github.com/actiontech/dms/pkg/params"
)

const (
	redisConnectionModeParam      = "connection_mode"
	redisConnectionModeStandalone = "standalone"
	redisConnectionModeCluster    = "cluster"
)

func isRedisDBType(dbType string) bool {
	return strings.EqualFold(dbType, string(pkgConst.DBTypeRedis))
}

func normalizeRedisConnectionModeValue(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", redisConnectionModeStandalone:
		return redisConnectionModeStandalone, nil
	case redisConnectionModeCluster:
		return redisConnectionModeCluster, nil
	default:
		return "", fmt.Errorf("invalid redis connection_mode: %s", value)
	}
}

func isRedisClusterDBServiceArgs(args *BizDBServiceArgs) bool {
	if args == nil || !isRedisDBType(args.DBType) {
		return false
	}
	mode, err := normalizeRedisConnectionModeValue(args.AdditionalParams.GetParam(redisConnectionModeParam).String())
	return err == nil && mode == redisConnectionModeCluster
}

func appendRedisConnectionModeIfMissing(dbType string, additionalParams []*dmsCommonV1.AdditionalParam) ([]*dmsCommonV1.AdditionalParam, error) {
	if !isRedisDBType(dbType) {
		return additionalParams, nil
	}
	mode := ""
	for _, item := range additionalParams {
		if item != nil && item.Name == redisConnectionModeParam {
			mode = item.Value
			break
		}
	}
	normalizedMode, err := normalizeRedisConnectionModeValue(mode)
	if err != nil {
		return nil, err
	}
	for _, item := range additionalParams {
		if item != nil && item.Name == redisConnectionModeParam {
			item.Value = normalizedMode
			return additionalParams, nil
		}
	}
	return append(additionalParams, &dmsCommonV1.AdditionalParam{
		Name:  redisConnectionModeParam,
		Value: normalizedMode,
		Type:  string(pkgParams.ParamTypeString),
	}), nil
}
