package service

import (
	"fmt"
	"strings"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/params"
)

const (
	redisConnectionModeParam      = "connection_mode"
	redisConnectionModeStandalone = "standalone"
	redisConnectionModeCluster    = "cluster"
)

func isRedisDBType(dbType string) bool {
	return strings.EqualFold(dbType, string(pkgConst.DBTypeRedis))
}

func isRedisConnectionModeParam(dbType, name string) bool {
	return isRedisDBType(dbType) && name == redisConnectionModeParam
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

func setRedisConnectionModeParam(additionalParams *params.Params, value string) error {
	mode, err := normalizeRedisConnectionModeValue(value)
	if err != nil {
		return err
	}
	if additionalParams == nil {
		return fmt.Errorf("additional params is nil")
	}
	if param := additionalParams.GetParam(redisConnectionModeParam); param != nil {
		param.Value = mode
		return nil
	}
	*additionalParams = append(*additionalParams, &params.Param{
		Key:   redisConnectionModeParam,
		Value: mode,
		Desc:  "Redis connection mode",
		Type:  params.ParamTypeString,
	})
	return nil
}

func normalizeRedisConnectionModeParams(dbType string, additionalParams *params.Params) error {
	if !isRedisDBType(dbType) {
		return nil
	}
	value := ""
	if additionalParams != nil {
		value = additionalParams.GetParam(redisConnectionModeParam).String()
	}
	return setRedisConnectionModeParam(additionalParams, value)
}

func validateDBServiceUser(dbType, user string, additionalParams params.Params) error {
	if isRedisDBType(dbType) {
		mode, err := normalizeRedisConnectionModeValue(additionalParams.GetParam(redisConnectionModeParam).String())
		if err != nil {
			return err
		}
		if mode == redisConnectionModeCluster {
			return nil
		}
	}
	if user == "" {
		return fmt.Errorf("db service user can't be empty")
	}
	return nil
}

func normalizeCheckDbConnectable(dbService *dmsCommonV1.CheckDbConnectable) error {
	if dbService == nil || !isRedisDBType(dbService.DBType) {
		if dbService != nil && dbService.User == "" {
			return fmt.Errorf("db service user can't be empty")
		}
		return nil
	}
	mode := ""
	for _, item := range dbService.AdditionalParams {
		if item != nil && item.Name == redisConnectionModeParam {
			mode = item.Value
			break
		}
	}
	normalizedMode, err := normalizeRedisConnectionModeValue(mode)
	if err != nil {
		return err
	}
	found := false
	for _, item := range dbService.AdditionalParams {
		if item != nil && item.Name == redisConnectionModeParam {
			item.Value = normalizedMode
			found = true
			break
		}
	}
	if !found {
		dbService.AdditionalParams = append(dbService.AdditionalParams, &dmsCommonV1.AdditionalParam{
			Name:  redisConnectionModeParam,
			Value: normalizedMode,
			Type:  string(params.ParamTypeString),
		})
	}
	if normalizedMode != redisConnectionModeCluster && dbService.User == "" {
		return fmt.Errorf("db service user can't be empty")
	}
	return nil
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
		Name:        redisConnectionModeParam,
		Value:       normalizedMode,
		Description: "Redis connection mode",
		Type:        string(params.ParamTypeString),
	}), nil
}
