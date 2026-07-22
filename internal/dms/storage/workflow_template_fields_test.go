package storage

import (
	"testing"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/stretchr/testify/assert"
)

func TestConvertWorkflowTemplateFields(t *testing.T) {
	now := time.Now()
	b := &biz.Workflow{
		UID:                  "uid-1",
		Name:                 "wf",
		ProjectUID:           "p1",
		WorkflowType:         "data_export",
		CreateTime:           now,
		CreateUserUID:        "u1",
		WorkflowRecordUid:    "r1",
		WorkflowTemplateId:   15,
		WorkflowTemplateName: "export-tmpl",
		WorkflowRecord: &biz.WorkflowRecord{
			UID:    "r1",
			Status: biz.DataExportWorkflowStatusWaitForApprove,
			Tasks:  []biz.Task{{UID: "t1"}},
		},
	}
	m := convertBizWorkflow(b)
	assert.Equal(t, uint(15), m.WorkflowTemplateId)
	assert.Equal(t, "export-tmpl", m.WorkflowTemplateName)

	got, err := convertModelWorkflow(m)
	assert.NoError(t, err)
	assert.Equal(t, uint(15), got.WorkflowTemplateId)
	assert.Equal(t, "export-tmpl", got.WorkflowTemplateName)
}
