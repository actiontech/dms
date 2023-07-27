package process_queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

var defaultGcSleepTime = 60 * time.Second

type processRecord struct {
	User       string
	Request    string
	HttpCode   int
	StartTime  time.Time
	FinishTime time.Time
	ErrorCode  string
	ErrorMsg   string
}

type ProcessRecordQueue struct {
	records   []*processRecord
	log       *log.Helper
	gcSeconds int // how long Queue keep finished records
	sync.Mutex
	PersistenceOperationRepo
}

func NewProcessRecord(user string, request string) *processRecord {
	return &processRecord{
		User:    user,
		Request: request,
	}
}

func NewProcessRecordQueue(logger log.Logger, recordFilePath string) *ProcessRecordQueue {
	persistenceRepo := NewProcessPersistenceRepo(logger, recordFilePath)

	p := &ProcessRecordQueue{
		records:                  make([]*processRecord, 0),
		gcSeconds:                365 * 24 * 3600, // default gc time: 1 year
		log:                      log.NewHelper(log.With(logger, "pkg", "process_queue")),
		PersistenceOperationRepo: persistenceRepo,
	}

	processRecords, err := persistenceRepo.ReadAll()
	if err != nil {
		log.Errorf("failed to read queues from persistence file %v", err.Error())
	} else {
		p.records = processRecords
	}

	go func() {
		for {
			time.Sleep(defaultGcSleepTime)
			p.gc()
		}
	}()
	return p
}

func (u *ProcessRecordQueue) gc() {
	u.Lock()
	defer u.Unlock()

	if len(u.records) == 0 {
		return
	}

	needSavedTime := time.Duration(u.gcSeconds) * time.Second
	firstRecordDurationTime := time.Since(u.records[0].StartTime)

	// 因为 processRecord.StartTime 只会递增 , 所以如果第一条记录存在的时间小于 needSavedTime , 则不需要清理
	// 例：假设日志需要保存一年，但所有的日志记录都在一年内发生，不做操作，直接 return
	if firstRecordDurationTime < needSavedTime {
		return
	}

	var needSavedRecords []*processRecord
	for _, record := range u.records {
		durationTime := time.Since(record.StartTime)
		if durationTime < needSavedTime {
			needSavedRecords = append(needSavedRecords, record)
		}
	}

	u.updateAndWrite(needSavedRecords)
}

func (u *ProcessRecordQueue) SetGcSeconds(gcSeconds int) error {
	u.Lock()
	defer u.Unlock()
	if gcSeconds < 1 {
		return fmt.Errorf("gc seconds should more than 1")
	}

	u.gcSeconds = gcSeconds
	u.log.Infof("[ProcessRecordQueue.SetGcSeconds] %v", u.gcSeconds)
	return nil
}

func (u *ProcessRecordQueue) Push(user, url string) *processRecord {
	u.Lock()
	defer u.Unlock()

	record := NewProcessRecord(user, url)

	record.StartTime = time.Now()
	u.records = append(u.records, record)
	u.log.Infof("[ProcessRecordQueue.Push] %v pushed into queue", record.Request)

	if err := u.WriteAll(u.records); err != nil {
		u.log.Errorf("failed to write queues to persistence file %v", err.Error())
	}

	return record
}

func (u *ProcessRecordQueue) List() []*processRecord {
	u.Lock()
	defer u.Unlock()

	ret := make([]*processRecord, len(u.records))

	copy(ret, u.records)
	return ret
}

func (u *ProcessRecordQueue) Save() {
	u.Lock()
	defer u.Unlock()

	if err := u.WriteAll(u.records); err != nil {
		u.log.Errorf("failed to write queues to persistence file %v", err.Error())
	}
}

func (u *ProcessRecordQueue) updateAndWrite(processRecords []*processRecord) {
	u.records = processRecords

	if err := u.PersistenceOperationRepo.WriteAll(processRecords); err != nil {
		u.log.Errorf("failed to write queues to persistence file %v", err.Error())
	}
}

func (p *processRecord) UpdateRecord(httpCode int, errorCode string, errorMsg string) {
	p.HttpCode = httpCode
	p.FinishTime = time.Now()
	p.ErrorCode = errorCode
	p.ErrorMsg = errorMsg
}
