//go:build dms

package biz

import (
	"context"

	maskBiz "github.com/actiontech/dms/internal/data_masking/biz"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
)

func (d *DataMaskingUsecase) ListMaskingRules(ctx context.Context) ([]ListMaskingRule, error) {
	rules, err := d.DataMasking.GetMaskingRulesOut()
	if err != nil {
		return nil, err
	}

	ret := make([]ListMaskingRule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, ListMaskingRule{
			MaskingType:     rule.MaskingType,
			Description:     rule.Description,
			ReferenceFields: rule.ReferenceFields,
			Effect:          rule.Effect,
		})
	}

	return ret, nil
}

// SQLExecuteResultsDataMasking 为DMS企业版的脱敏功能，捕获cloudbeaver返回的结果集，根据配置对结果集脱敏
func (d *DataMaskingUsecase) SQLExecuteResultsDataMasking(ctx context.Context, result *model.SQLExecuteInfo) error {
	for _, r := range result.Results {
		if r.ResultSet == nil {
			continue
		}
		c := make([]*maskBiz.SqlResultColumn, len(r.ResultSet.Columns))
		for i := range r.ResultSet.Columns {
			c[i] = &maskBiz.SqlResultColumn{
				Name: *r.ResultSet.Columns[i].Name,
			}
		}

		params := maskBiz.NewMaskSqlExecuteResultParams(c)
		for i := range r.ResultSet.Rows {
			params.AddRows(r.ResultSet.Rows[i])
		}

		if err := d.DataMasking.MaskSqlExecuteResultByAutoDetection(params); nil != err {
			return err
		}
	}

	return nil
}

func IsDMS() bool {
	return true
}
