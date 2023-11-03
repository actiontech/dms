package process_queue

import (
	"encoding/json"
	"fmt"
	"sync"

	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	utilIo "github.com/actiontech/dms/pkg/dms-common/pkg/io"

	"github.com/go-kratos/kratos/v2/log"
)

var _ PersistenceOperationRepo = (*ProcessPersistenceRepo)(nil)

type PersistenceOperationRepo interface {
	ReadAll() ([]*processRecord, error)
	WriteAll([]*processRecord) error
}

type ProcessPersistenceRepo struct {
	log log.Logger
	sync.Mutex
	ProcessOperationLog string
}

func (p *ProcessPersistenceRepo) ReadAll() (processRecords []*processRecord, err error) {
	p.Lock()
	defer p.Unlock()

	if !utilIo.IsFileExist(p.ProcessOperationLog) {
		return processRecords, nil
	}

	bytes, err := utilIo.ReadFile(pkgLog.NewUtilLogWrapper(p.log), p.ProcessOperationLog)
	if err != nil {
		return nil, fmt.Errorf("failed to read process operation log: %v", err)
	}

	if err = json.Unmarshal(bytes, &processRecords); err != nil {
		return nil, fmt.Errorf("failed to process unmarshal: %v", err)
	}

	return processRecords, nil
}

func (p *ProcessPersistenceRepo) WriteAll(processRecords []*processRecord) error {
	p.Lock()
	defer p.Unlock()

	bytes, err := json.Marshal(processRecords)
	if err != nil {
		return fmt.Errorf("failed to process marshal: %v", err)
	}

	if err = utilIo.WriteFile(pkgLog.NewUtilLogWrapper(p.log), p.ProcessOperationLog, string(bytes), "", 0640); err != nil {
		return fmt.Errorf("failed to write process operation log: %v", err)
	}

	return nil
}

func NewProcessPersistenceRepo(log log.Logger, persistenceFilePath string) PersistenceOperationRepo {
	return &ProcessPersistenceRepo{
		log:                 log,
		ProcessOperationLog: persistenceFilePath,
	}
}
