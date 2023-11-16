//go:build enterprise

package biz

import (
	"context"
	"errors"

	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
)

func (d *CompanyNoticeUsecase) UpdateCompanyNotice(ctx context.Context, noticeStr *string) error {
	notice, err := d.repo.GetCompanyNotice(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到notice,初始化一个带uid的值
		notice, err = initCompanyNotice()
		if err != nil {
			return err
		}
	}

	// patch notice str
	if noticeStr != nil {
		notice.NoticeStr = *noticeStr
		notice.ReadUserIds = []string{}
	}

	return d.repo.UpdateCompanyNotice(ctx, notice)
}

func (d *CompanyNoticeUsecase) GetCompanyNotice(ctx context.Context, userId string) (notice *CompanyNotice, read bool, err error) {
	notice, err = d.repo.GetCompanyNotice(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	user, err := d.userUsecase.GetUser(ctx, userId)
	if err != nil {
		return nil, false, err
	}
	for _, userId := range notice.ReadUserIds {
		if user.UID == userId {
			return notice, true, nil
		}
	}
	// update user read record
	notice.ReadUserIds = append(notice.ReadUserIds, userId)
	err = d.repo.UpdateCompanyNotice(ctx, notice)
	if err != nil {
		return notice, false, err
	}
	return notice, false, nil
}
