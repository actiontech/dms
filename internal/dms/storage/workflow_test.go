package storage

import (
	"context"
	"testing"
	"time"

	"github.com/actiontech/dms/internal/dms/storage/model"
)

func TestListExpiredWorkflowUIDs(t *testing.T) {
	storage := GetTestStorage(t)
	repo := NewWorkflowRepo(testLogger, storage)
	now := time.Now()
	expiredWorkflowUID := "expired-recycle-test"
	activeWorkflowUID := "active-recycle-test"
	workflows := []*model.Workflow{
		{
			Model:      model.Model{UID: expiredWorkflowUID},
			Name:       expiredWorkflowUID,
			ProjectUID: "recycle-test-project",
			CreateTime: timePointer(now.Add(-366 * 24 * time.Hour)),
		},
		{
			Model:      model.Model{UID: activeWorkflowUID},
			Name:       activeWorkflowUID,
			ProjectUID: "recycle-test-project",
			CreateTime: timePointer(now),
		},
	}

	if err := storage.db.Create(workflows).Error; err != nil {
		t.Fatalf("create workflows: %v", err)
	}
	t.Cleanup(func() {
		storage.db.Unscoped().Delete(&model.Workflow{}, "uid IN ?", []string{expiredWorkflowUID, activeWorkflowUID})
	})

	workflowUIDs, err := repo.ListExpiredWorkflowUIDs(context.Background(), now.Add(-365*24*time.Hour))
	if err != nil {
		t.Fatalf("list expired workflow UIDs: %v", err)
	}
	if !containsString(workflowUIDs, expiredWorkflowUID) {
		t.Fatalf("expected expired workflow UID %q in %v", expiredWorkflowUID, workflowUIDs)
	}
	if containsString(workflowUIDs, activeWorkflowUID) {
		t.Fatalf("did not expect active workflow UID %q in %v", activeWorkflowUID, workflowUIDs)
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
