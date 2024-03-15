package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListMaskingRules
type ListMaskingRulesReq struct {
}

type ListMaskingRulesData struct {
	MaskingType     string   `json:"masking_type"`
	Description     string   `json:"description"`
	ReferenceFields []string `json:"reference_fields"`
	Effect          string   `json:"effect"`
	Id              int      `json:"id"`
}

// swagger:model ListMaskingRulesReply
type ListMaskingRulesReply struct {
	// list masking rule reply
	Data []ListMaskingRulesData `json:"data"`

	// Generic reply
	base.GenericResp
}
