package v1

import base "github.com/actiontech/dms/api/base/v1"

// swagger:model GetSQLQueryConfigurationReply
type GetSQLQueryConfigurationReply struct {
	Payload struct {
		EnableSQLQuery  bool   `json:"enable_sql_query"`
		SQLQueryRootURI string `json:"sql_query_root_uri"`
	} `json:"payload"`
	// Generic reply
	base.GenericResp
}
