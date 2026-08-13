package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

type AccessWhitelistRuleItem struct {
	UID        string `json:"uid"`
	Source     string `json:"source"`
	PolicyType string `json:"policy_type"`
	Remark     string `json:"remark"`
	UpdatedAt  string `json:"updated_at"`
}

type AccessRestrictionConfig struct {
	Enabled bool                      `json:"enabled"`
	Rules   []AccessWhitelistRuleItem `json:"rules"`
}

// swagger:model GetAccessRestrictionReply
type GetAccessRestrictionReply struct {
	Data AccessRestrictionConfig `json:"data"`
	base.GenericResp
}

// swagger:model
type UpdateAccessRestrictionReq struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

// swagger:model
type CreateAccessWhitelistRuleReq struct {
	Source     string `json:"source" validate:"required"`
	Remark     string `json:"remark"`
	PolicyType string `json:"policy_type"`
}

// swagger:model CreateAccessWhitelistRuleReply
type CreateAccessWhitelistRuleReply struct {
	Data AccessWhitelistRuleItem `json:"data"`
	base.GenericResp
}

// swagger:parameters UpdateAccessWhitelistRuleReq
type UpdateAccessWhitelistRuleReq struct {
	// in:path
	RuleUID    string `param:"rule_uid" json:"rule_uid" validate:"required"`
	Source     string `json:"source" validate:"required"`
	Remark     string `json:"remark"`
	PolicyType string `json:"policy_type"`
}

// swagger:model UpdateAccessWhitelistRuleReply
type UpdateAccessWhitelistRuleReply struct {
	Data AccessWhitelistRuleItem `json:"data"`
	base.GenericResp
}

// swagger:parameters DeleteAccessWhitelistRuleReq
type DeleteAccessWhitelistRuleReq struct {
	// in:path
	RuleUID string `param:"rule_uid" json:"rule_uid" validate:"required"`
}

// swagger:model GetAccessRestrictionClientIPReply
type GetAccessRestrictionClientIPReply struct {
	Data AccessRestrictionClientIP `json:"data"`
	base.GenericResp
}

type AccessRestrictionClientIP struct {
	ClientIP string `json:"client_ip"`
}
