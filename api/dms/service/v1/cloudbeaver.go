package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:model GetSQLQueryConfigurationReply
type GetSQLQueryConfigurationReply struct {
	Data struct {
		EnableSQLQuery  bool   `json:"enable_sql_query"`
		SQLQueryRootURI string `json:"sql_query_root_uri"`
	} `json:"data"`
	// Generic reply
	base.GenericResp
}
