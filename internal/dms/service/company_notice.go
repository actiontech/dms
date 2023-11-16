package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) GetCompanyNotice(ctx context.Context, currentUserUid string) (reply *dmsV1.GetCompanyNoticeReply, err error) {
	companyNotice, exist, err := d.CompanyNoticeUsecase.GetCompanyNotice(ctx, currentUserUid)
	if err != nil {
		return nil, err
	}

	data := dmsV1.CompanyNotice{}
	if exist {
		data.NoticeStr = companyNotice.NoticeStr
		data.ReadByCurrentUser = companyNotice.ReadByCurrentUser
	}
	return &dmsV1.GetCompanyNoticeReply{
		Data: data,
	}, nil
}

func (d *DMSService) UpdateCompanyNotice(ctx context.Context, req *dmsV1.UpdateCompanyNoticeReq) (err error) {
	return d.CompanyNoticeUsecase.UpdateCompanyNotice(ctx, req.UpdateCompanyNotice.NoticeStr)
}
