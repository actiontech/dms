package storage

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	kLog "github.com/go-kratos/kratos/v2/log"
)

const (
	TestMySQLPort     = 33306
	TestMySQLHost     = "127.0.0.1"
	TestMySQLUser     = "root"
	TestMySQLPassword = "123"
	TestMySQLSchema   = "dms_unittest"
)

var testLogger = pkgLog.NewUtilLogWrapper(kLog.With(pkgLog.NewStdLogger(os.Stdout, pkgLog.LogTimeLayout),
	"caller", kLog.DefaultCaller,
))

var testStorage *Storage
var InitTestStorageOnce sync.Once

func GetTestStorage(t *testing.T) (s *Storage) {
	if err := waitUntilTestDBPortOpen(); nil != err {
		t.Fatalf("wait until test db port open error: %v", err)
	}
	InitTestStorageOnce.Do(func() {
		var err error
		testStorage, err = NewStorage(testLogger, &StorageConfig{
			Host:     TestMySQLHost,
			Port:     fmt.Sprintf("%d", TestMySQLPort),
			User:     TestMySQLUser,
			Password: TestMySQLPassword,
			Schema:   TestMySQLSchema,
		})
		if nil != err {
			t.Fatalf("new orm storage error: %v", err)
		}
		cleanupTestStorage()
		err = TestInitStorage(testStorage)
		if nil != err {
			t.Fatalf("init test storage error: %v", err)
		}
	})

	return testStorage
}

func waitUntilTestDBPortOpen() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(TestMySQLHost, fmt.Sprintf("%v", TestMySQLPort)), 1*time.Minute)
	if nil != err {
		return fmt.Errorf("wait until test db port open error: %v", err)
	}
	if conn == nil {
		return fmt.Errorf("wait until test db port open error: no conn")
	}
	conn.Close()
	return nil
}

func TestInitStorage(s *Storage) (err error) {
	err = s.db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v", TestMySQLSchema)).Error
	if nil != err {
		return err
	}
	return s.AutoMigrate(testLogger)
}

func cleanupTestStorage() {
	if testStorage == nil {
		fmt.Println("skip cleanup db")
	}
	err := testStorage.db.Exec(fmt.Sprintf("DROP DATABASE IF EXIST %v", TestMySQLSchema)).Error
	if nil != err {
		fmt.Println("clean up db errors: ", err)
		return
	}
	fmt.Println("clean up db")
}
