package service

import (
	"context"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
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
		Payload: struct {
			BasicInfo *v1.BasicInfo `json:"basic_info"`
		}{ret},
	}, nil
}
