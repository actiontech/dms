package service

import (
	"testing"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/params"
)

func TestValidateDBServiceUser_RedisAllowsEmpty(t *testing.T) {
	t.Parallel()
	redisParams := params.Params{
		{Key: redisConnectionModeParam, Value: redisConnectionModeStandalone, Type: params.ParamTypeString},
	}
	if err := validateDBServiceUser("Redis", "", redisParams); err != nil {
		t.Fatalf("redis standalone empty user: %v", err)
	}
	clusterParams := params.Params{
		{Key: redisConnectionModeParam, Value: redisConnectionModeCluster, Type: params.ParamTypeString},
	}
	if err := validateDBServiceUser("Redis", "", clusterParams); err != nil {
		t.Fatalf("redis cluster empty user: %v", err)
	}
}

func TestValidateDBServiceUser_NonRedisRejectsEmpty(t *testing.T) {
	t.Parallel()
	if err := validateDBServiceUser("MySQL", "", nil); err == nil {
		t.Fatal("expected non-redis empty user rejected")
	}
}

func TestNormalizeCheckDbConnectable_RedisStandaloneEmptyUser(t *testing.T) {
	t.Parallel()
	args := &dmsCommonV1.CheckDbConnectable{
		DBType: "Redis",
		User:   "",
		AdditionalParams: []*dmsCommonV1.AdditionalParam{
			{Name: redisConnectionModeParam, Value: redisConnectionModeStandalone},
		},
	}
	if err := normalizeCheckDbConnectable(args); err != nil {
		t.Fatalf("redis standalone empty user connectable: %v", err)
	}
}

func TestNormalizeCheckDbConnectable_MySQLEmptyUser(t *testing.T) {
	t.Parallel()
	args := &dmsCommonV1.CheckDbConnectable{DBType: "MySQL", User: ""}
	if err := normalizeCheckDbConnectable(args); err == nil {
		t.Fatal("expected mysql empty user rejected")
	}
}
