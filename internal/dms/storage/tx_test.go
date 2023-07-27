package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/actiontech/dms/internal/dms/biz"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

func Test_TXCommit(t *testing.T) {
	log := utilLog.NewHelper(testLogger, utilLog.WithMessageKey("unittest.Test_TX"))
	txGen := NewTXGenerator()
	s := GetTestStorage(t)
	userRepo := NewUserRepo(testLogger, s)

	tx := txGen.BeginTX(context.Background())

	if err := userRepo.SaveUser(tx, &biz.User{
		UID:  "2023032201",
		Name: "2023032201",
	}); err != nil {
		t.Fatalf("save user error: %v", err)
	}
	if err := tx.Commit(log); err != nil {
		t.Fatalf("commit tx error: %v", err)
	}

	exist, err := userRepo.CheckUserExist(context.Background(), []string{"2023032201"})
	if err != nil {
		t.Fatalf("check user exist error: %v", err)
	}
	if !exist {
		t.Fatalf("check user exist error: user should exist")
	}

}

func Test_TXRollback(t *testing.T) {
	log := utilLog.NewHelper(testLogger, utilLog.WithMessageKey("unittest.Test_TXRollback"))
	txGen := NewTXGenerator()
	s := GetTestStorage(t)
	userRepo := NewUserRepo(testLogger, s)

	tx := txGen.BeginTX(context.Background())

	if err := userRepo.SaveUser(tx, &biz.User{
		UID:  "2023032202",
		Name: "2023032202",
	}); err != nil {
		t.Fatalf("save user error: %v", err)
	}
	if err := tx.RollbackWithError(log, fmt.Errorf("testerr")); err == nil {
		t.Fatalf("rollback tx error: error should be `testerr`, but got nil")
	} else if err.Error() != "testerr" {
		t.Fatalf("rollback tx error: error should be `testerr`, but got %v", err)
	}

	exist, err := userRepo.CheckUserExist(context.Background(), []string{"2023032202"})
	if err != nil {
		t.Fatalf("check user exist error: %v", err)
	}
	if exist {
		t.Fatalf("check user exist error: user should not exist")
	}
}
