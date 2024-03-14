//go:build enterprise

package biz

import (
	_ "embed"
	"fmt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	dlp "github.com/bytedance/godlp"
	"github.com/bytedance/godlp/conf"
	"github.com/bytedance/godlp/dlpheader"
	"gopkg.in/yaml.v2"
)

//go:embed data_masking_conf_ee.yml
var ruleConf string

// 数据脱敏
type DataMaskingUseCase struct {
	log *utilLog.Helper
	eng dlpheader.EngineAPI
}

func NewDataMaskingUseCase(log utilLog.Logger) (*DataMaskingUseCase, error) {
	eng, err := dlp.NewEngine("dms.data_masking")
	if nil != err {
		return nil, fmt.Errorf("failed to new data masking engine: %v", err)
	}
	// TODO: 用户可配置规则
	if err := eng.ApplyConfig(ruleConf); nil != err {
		return nil, fmt.Errorf("failed to apply data masking config: %v", err)
	}
	d := &DataMaskingUseCase{
		log: utilLog.NewHelper(log, utilLog.WithMessageKey("data_masking")),
		eng: eng,
	}
	return d, nil
}

type MaskSqlExecuteResultParams struct {
	ResultSet *SqlResultSet
}

type SqlResultSet struct {
	Columns []*SqlResultColumn
	Rows    [][]interface{} `json:"rows"`
}

type SqlResultColumn struct {
	Name string `json:"name"`
	// TODO: more info
}

func NewMaskSqlExecuteResultParams(c []*SqlResultColumn) *MaskSqlExecuteResultParams {
	return &MaskSqlExecuteResultParams{
		ResultSet: &SqlResultSet{
			Columns: c,
		},
	}
}

func (d *MaskSqlExecuteResultParams) AddRows(rows []interface{}) error {
	if len(rows) != len(d.ResultSet.Columns) {
		return fmt.Errorf("columns count %v is not equal to rows count %v", len(d.ResultSet.Columns), len(rows))
	}
	d.ResultSet.Rows = append(d.ResultSet.Rows, rows)
	return nil
}

func (d *MaskSqlExecuteResultParams) HasRows() bool {
	return len(d.ResultSet.Rows) > 0
}

func (d *DataMaskingUseCase) MaskSqlExecuteResultByAutoDetection(data *MaskSqlExecuteResultParams) error {
	if !data.HasRows() {
		return nil
	}
	if data.ResultSet == nil {
		return nil
	}

	for ri, r := range data.ResultSet.Rows {
		if len(data.ResultSet.Columns) != len(r) {
			return fmt.Errorf("columns count %v is not equal to rows count %v", len(data.ResultSet.Columns), len(r))
		}

		for i, v := range r {
			value := fmt.Sprintf("%v", v)
			columnName := data.ResultSet.Columns[i].Name
			m := map[string]string{
				columnName: value,
			}

			out, results, err := d.eng.DeidentifyMap(m)
			if nil != err {
				return fmt.Errorf("failed to deidentify map: %v", err)
			}
			if len(results) == 1 {
				data.ResultSet.Rows[ri][i] = out[columnName]
			}
		}
	}
	return nil
}

//go:embed data_masking_rule_in.yml
var dataMaskingRuleInConf string

type DataMaskingRuleIn struct {
	Items []DataMaskingRuleInItem `yaml:"Items"`
}

type DataMaskingRuleInItem struct {
	RuleID int32  `yaml:"RuleID"`
	In     string `yaml:"In"`
	Method string `yaml:"Method"`
}

type DataMaskingRuleOut struct {
	MaskingType     string   `json:"masking_type"`
	ReferenceFields []string `json:"reference_fields"`
	Effect          string   `json:"effect"`
}

func (d *DataMaskingUseCase) GetMaskingRulesOut() ([]DataMaskingRuleOut, error) {
	var dataMaskingRuleIn DataMaskingRuleIn
	ruleIdOutMap := make(map[int32]string)
	if err := yaml.Unmarshal([]byte(dataMaskingRuleInConf), &dataMaskingRuleIn); err == nil {
		for _, item := range dataMaskingRuleIn.Items {
			if out, err := d.eng.Mask(item.In, item.Method); err == nil {
				ruleIdOutMap[item.RuleID] = out
			} else {
				d.log.Errorf("masking out err: %v", err)
			}
		}
	} else {
		d.log.Errorf("masking rule conf unmarshal err: %v", err)
	}

	dataMaskingConf := conf.DlpConf{}

	if err := yaml.Unmarshal([]byte(d.eng.GetDefaultConf()), &dataMaskingConf); err != nil {
		return nil, err
	}

	ret := make([]DataMaskingRuleOut, 0, len(dataMaskingConf.Rules))
	for _, rule := range dataMaskingConf.Rules {
		ret = append(ret, DataMaskingRuleOut{
			MaskingType:     fmt.Sprintf("%s (%s)", rule.CnName, rule.Description),
			ReferenceFields: rule.Verify.CDict,
			Effect:          ruleIdOutMap[rule.RuleID],
		})
	}

	return ret, nil
}
