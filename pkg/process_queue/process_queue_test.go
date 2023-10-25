package process_queue

import (
	"os"
	"testing"
	"time"

	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	utilIo "github.com/actiontech/dms/pkg/dms-common/pkg/io"

	"github.com/go-kratos/kratos/v2/log"
)

var processRecordTestRequest1 = "test-request-1"
var processRecordTestRequest2 = "test-request-2"
var processRecordTestRequest3 = "test-request-3"
var user1 = "u1"
var user2 = "u2"

var processRecordTestFailResultCode1 = 500

var testLogger = log.With(log.NewStdLogger(os.Stdout),
	"service.name", "process-queue-unit-test",
	"ts", log.DefaultTimestamp,
	"caller", log.DefaultCaller,
)

var testRecordFilePath = "test-record-file-path"

func TestProcessRecordQueue_GcRemoveExpiredOperationLog(t *testing.T) {
	defaultGcSleepTime = 2 * time.Second

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)

	record1 := q.Push(user1, processRecordTestRequest1)
	record2 := q.Push(user1, processRecordTestRequest2)
	record3 := q.Push(user2, processRecordTestRequest3)

	record1.UpdateRecord(500, "0", "msg")
	record1.StartTime = record1.StartTime.Add(-2 * 365 * 24 * time.Hour)
	record2.UpdateRecord(500, "0", "msg")
	record2.StartTime = record2.StartTime.Add(-400 * 24 * time.Hour)
	record3.UpdateRecord(404, "0", "msg3")
	record3.StartTime = record3.StartTime.Add(100 * 24 * time.Hour)

	time.Sleep(3 * time.Second)

	if len(q.records) != 1 {
		t.Fatalf(`process record queue should have 1 length but have %v`, len(q.List()))
	}

	if q.records[0].User != "u2" {
		t.Fatalf(`process record queue should have u2 but have %v`, q.records[0].User)
	}

	if q.records[0].HttpCode != 404 {
		t.Fatalf(`process record queue should have 404 http code but have %v`, q.records[0].HttpCode)
	}

	if q.records[0].ErrorMsg != "msg3" {
		t.Fatalf(`process record queue should have msg3 error msg but have %v`, q.records[0].ErrorMsg)
	}

	defaultGcSleepTime = 60 * time.Second
}

func TestProcessRecordQueue_AllOperationLogStartTimeLessOneYear(t *testing.T) {
	defaultGcSleepTime = 2 * time.Second

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)

	record1 := q.Push(user1, processRecordTestRequest1)
	record2 := q.Push(user1, processRecordTestRequest2)
	record3 := q.Push(user2, processRecordTestRequest3)

	record1.UpdateRecord(500, "0", "msg")
	record1.StartTime = record1.StartTime.Add(-10 * 24 * time.Hour)
	record2.UpdateRecord(500, "0", "msg")
	record2.StartTime = record2.StartTime.Add(-20 * 24 * time.Hour)
	record3.UpdateRecord(404, "0", "msg3")
	record3.StartTime = record3.StartTime.Add(-15 * 24 * time.Hour)

	time.Sleep(3 * time.Second)

	if len(q.records) != 3 {
		t.Fatalf(`process record queue should have 3 length but have %v`, len(q.List()))
	}

	defaultGcSleepTime = 60 * time.Second
}

func TestProcessRecordQueue_AllOperationLogStartTimeMoreThenOneYear(t *testing.T) {
	defaultGcSleepTime = 2 * time.Second

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)

	record1 := q.Push(user1, processRecordTestRequest1)
	record2 := q.Push(user1, processRecordTestRequest2)
	record3 := q.Push(user2, processRecordTestRequest3)

	record1.UpdateRecord(500, "0", "msg")
	record1.StartTime = record1.StartTime.Add(-20 * 365 * 24 * time.Hour)
	record2.UpdateRecord(500, "0", "msg")
	record2.StartTime = record2.StartTime.Add(-15 * 365 * 24 * time.Hour)
	record3.UpdateRecord(404, "0", "msg3")
	record3.StartTime = record3.StartTime.Add(-10 * 365 * 24 * time.Hour)

	time.Sleep(3 * time.Second)

	if len(q.records) != 0 {
		t.Fatalf(`process record queue should have 0 length but have %v`, len(q.List()))
	}

	defaultGcSleepTime = 60 * time.Second
}

func TestProcessRecordQueue_SuccessResult(t *testing.T) {

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)
	record := q.Push(user1, processRecordTestRequest1)

	record.UpdateRecord(200, "8000", "error_message")

	resultRecords := q.List()
	if len(resultRecords) != 1 {
		t.Fatalf(`process record queue should have 1 length but have %v`, len(resultRecords))
	}
	if resultRecords[0].User != user1 {
		t.Fatalf(`process record user should be %v but %v`, user1, resultRecords[0].User)
	}
	if resultRecords[0].Request != processRecordTestRequest1 {
		t.Fatalf(`process record request should be %v but be %v`, processRecordTestRequest1, resultRecords[0].Request)
	}
	if resultRecords[0].HttpCode != 200 {
		t.Fatalf(`process record result should be 200 but be %v`, resultRecords[0].HttpCode)
	}
	if resultRecords[0].ErrorCode != "8000" {
		t.Fatalf(`process record error code should be 8000 but be %v`, resultRecords[0].ErrorCode)
	}
	if resultRecords[0].ErrorMsg != "error_message" {
		t.Fatalf(`process record error message should be error_message but be %v`, resultRecords[0].ErrorMsg)
	}
}

func TestProcessRecordQueue_FailResult(t *testing.T) {

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)
	record := q.Push(user1, processRecordTestRequest1)

	record.UpdateRecord(processRecordTestFailResultCode1, "0", "errorMessage")

	resultRecords := q.List()
	if len(resultRecords) != 1 {
		t.Fatalf(`process record queue should have 1 length but have %v`, len(resultRecords))
	}
	if resultRecords[0].Request != processRecordTestRequest1 {
		t.Fatalf(`process record request should be %v but be %v`, processRecordTestRequest1, resultRecords[0].Request)
	}
	if resultRecords[0].HttpCode != processRecordTestFailResultCode1 {
		t.Fatalf(`process record result should be %v but be %v`, processRecordTestFailResultCode1, resultRecords[0].HttpCode)
	}
}

func TestProcessRecordQueue_GcNotDeleteNotFinishedRecord(t *testing.T) {

	if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
		t.Fatalf("remove process operation log failed: %v", err)
	}

	defer func() {
		if err := utilIo.Remove(pkgLog.NewUtilLogWrapper(testLogger), testRecordFilePath); err != nil {
			t.Fatalf("remove process operation log failed: %v", err)
		}
	}()

	q := NewProcessRecordQueue(testLogger, testRecordFilePath)
	_ = q.SetGcSeconds(1)
	record1 := q.Push(user1, processRecordTestRequest1)
	record2 := q.Push(user1, processRecordTestRequest2)

	record1.UpdateRecord(500, "0", "msg")
	record2.UpdateRecord(500, "0", "msg")

	time.Sleep(3 * time.Second)

	resultRecords := q.List()
	if len(resultRecords) != 2 {
		t.Fatalf(`process record queue should have 2 length but have %v`, len(resultRecords))
	}
}
