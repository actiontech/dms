package errors

import (
	"errors"
	"fmt"
	"strings"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var (
	ErrStorage       = errors.New("dms: storage error")
	ErrStorageNoData = errors.New("dms: storage no data")
	// prevent users from malicious password attempts
	ErrBeenBoundOrThePasswordIsWrong = errors.New("the platform user has been bound or the password is wrong")
)

func WrapStorageErr(log *utilLog.Helper, originalErr error) error {
	if originalErr == nil {
		return nil
	}
	if strings.Contains(originalErr.Error(), gorm.ErrRecordNotFound.Error()) {
		return WrapErrStorageNoData(log, originalErr)
	}
	err := fmt.Errorf("%w:%v", ErrStorage, originalErr)
	log.Errorf(err.Error())
	return err
}

func WrapErrStorageNoData(log *utilLog.Helper, originalErr error) error {
	err := fmt.Errorf("%w:%v", ErrStorageNoData, originalErr)
	log.Errorf(err.Error())
	return err
}
