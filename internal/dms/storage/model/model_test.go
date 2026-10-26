package model

import (
	"reflect"
	"testing"
)

func TestWorkflowGetTaskIds(t *testing.T) {
	t.Run("workflow record is not loaded", func(t *testing.T) {
		workflow := &Workflow{}

		if taskIds := workflow.GetTaskIds(); taskIds != nil {
			t.Fatalf("expected nil task IDs, got %v", taskIds)
		}
	})

	t.Run("workflow record contains task IDs", func(t *testing.T) {
		expectedTaskIds := Strings{"task-1", "task-2"}
		workflow := &Workflow{
			WorkflowRecord: &WorkflowRecord{TaskIds: expectedTaskIds},
		}

		if taskIds := workflow.GetTaskIds(); !reflect.DeepEqual(taskIds, expectedTaskIds) {
			t.Fatalf("expected task IDs %v, got %v", expectedTaskIds, taskIds)
		}
	})
}
