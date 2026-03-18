//go:build !enterprise

package biz

import (
	"context"
	"errors"
	"time"
)

var errNotSupportCompanyNotice = errors.New("company notice related functions are enterprise version functions")

func (d *CompanyNoticeUsecase) UpdateCompanyNotice(ctx context.Context, noticeStr *string, startTime, endTime *time.Time) error {

	return errNotSupportCompanyNotice
}

func (d *CompanyNoticeUsecase) GetCompanyNotice(ctx context.Context, userId string) (notice *CompanyNotice, err error) {
	return nil, errNotSupportCompanyNotice
}
