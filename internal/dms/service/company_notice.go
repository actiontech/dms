package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) GetCompanyNotice(ctx context.Context, currentUserUid string, includeLatestOutsidePeriod bool) (reply *dmsV1.GetCompanyNoticeReply, err error) {
	companyNotice, err := d.CompanyNoticeUsecase.GetCompanyNotice(ctx, currentUserUid, includeLatestOutsidePeriod)
	if err != nil {
		return nil, err
	}
	data := dmsV1.CompanyNotice{
		ReadByCurrentUser: false,
	}
	if companyNotice != nil {
		data.NoticeStr = companyNotice.NoticeStr
		data.StartTime = companyNotice.StartTime
		data.ExpireTime = companyNotice.EndTime
	}
	return &dmsV1.GetCompanyNoticeReply{
		Data: data,
	}, nil
}

func (d *DMSService) UpdateCompanyNotice(ctx context.Context, req *dmsV1.UpdateCompanyNoticeReq) (err error) {
	return d.CompanyNoticeUsecase.UpdateCompanyNotice(ctx, req.UpdateCompanyNotice.NoticeStr, req.UpdateCompanyNotice.StartTime, req.UpdateCompanyNotice.EndTime)
}
