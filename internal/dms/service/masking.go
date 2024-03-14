package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) ListMaskingRules(ctx context.Context) (reply *dmsV1.ListMaskingRulesReply, err error) {
	rules, err := d.DataMaskingUsecase.ListMaskingRules(ctx)
	if nil != err {
		return nil, err
	}

	ret := make([]dmsV1.ListMaskingRulesData, 0, len(rules))
	for i, rule := range rules {
		var fields = make([]string, 0)
		if rule.ReferenceFields != nil {
			fields = rule.ReferenceFields
		}

		ret = append(ret, dmsV1.ListMaskingRulesData{
			Id:              i + 1,
			MaskingType:     rule.MaskingType,
			ReferenceFields: fields,
			Effect:          rule.Effect,
		})
	}

	return &dmsV1.ListMaskingRulesReply{
		Data: ret,
	}, nil
}
