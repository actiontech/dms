package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) GetCompanyNotice(ctx context.Context, currentUserUid string) (reply *dmsV1.GetCompanyNoticeReply, err error) {
	companyNotice, read, err := d.CompanyNoticeUsecase.GetCompanyNotice(ctx, currentUserUid)
	if err != nil {
		return nil, err
	}
	data := dmsV1.CompanyNotice{
		ReadByCurrentUser: read,
	}
	if companyNotice != nil {
		data.NoticeStr = companyNotice.NoticeStr
	}
	return &dmsV1.GetCompanyNoticeReply{
		Data: data,
	}, nil
}

func (d *DMSService) UpdateCompanyNotice(ctx context.Context, req *dmsV1.UpdateCompanyNoticeReq) (err error) {
	return d.CompanyNoticeUsecase.UpdateCompanyNotice(ctx, req.UpdateCompanyNotice.NoticeStr)
}
