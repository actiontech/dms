package service

import (
	"context"
	"errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (d *DMSService) GetBasicInfo(ctx context.Context) (reply *v1.GetBasicInfoReply, err error) {
	basic, err := d.BasicUsecase.GetBasicInfo(ctx)
	if nil != err {
		return nil, err
	}

	ret := &v1.BasicInfo{
		LogoUrl: basic.LogoUrl,
		Title:   basic.Title,
	}

	for _, item := range basic.Components {
		ret.Components = append(ret.Components, v1.ComponentNameWithVersion{
			Name:    item.Name,
			Version: item.Version,
		})
	}

	return &v1.GetBasicInfoReply{
		Data: ret,
	}, nil
}

func (d *DMSService) GetStaticLogo(ctx context.Context) (*v1.GetStaticLogoReply, string, error) {
	basicConfig, contentType, err := d.BasicUsecase.GetStaticLogo(ctx)
	if nil != err {
		return nil, contentType, err
	}

	return &v1.GetStaticLogoReply{
		File: basicConfig.Logo,
	}, contentType, nil
}

func (d *DMSService) Personalization(ctx context.Context, req *v1.PersonalizationReq) error {
	if req.Title == "" && req.File == nil {
		return errors.New("one of the parameters title, logo is required")
	}

	params := &biz.BasicConfigParams{
		Title: req.Title,
		File:  req.File,
	}

	return d.BasicUsecase.Personalization(ctx, params)
}

func (d *DMSService) GetLimitAndOffset(pageIndex, pageSize uint32) (limit, offset int) {
	if pageIndex >= 1 {
		offset = int((pageIndex - 1) * pageSize)
	}
	return int(pageSize), offset
}
