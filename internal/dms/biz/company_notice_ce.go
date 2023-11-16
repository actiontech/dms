//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportCompanyNotice = errors.New("company notice related functions are enterprise version functions")

func (d *CompanyNoticeUsecase) UpdateCompanyNotice(ctx context.Context, noticeStr *string) error {

	return errNotSupportCompanyNotice
}

func (d *CompanyNoticeUsecase) GetCompanyNotice(ctx context.Context, userId string) (notice *CompanyNotice, exist bool, err error) {
	return nil, false, errNotSupportCompanyNotice
}
