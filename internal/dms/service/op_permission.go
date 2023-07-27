package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (d *DMSService) ListOpPermissions(ctx context.Context, req *dmsV1.ListOpPermissionReq) (reply *dmsV1.ListOpPermissionReply, err error) {
	d.log.Infof("ListOpPermissions.req=%v", req)
	defer func() {
		d.log.Infof("ListOpPermissions.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.OpPermissionField
	switch req.OrderBy {
	case dmsV1.OpPermissionOrderByName:
		orderBy = biz.OpPermissionFieldName
	default:
		orderBy = biz.OpPermissionFieldName
	}

	listOption := &biz.ListOpPermissionsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
	}

	var ops []*biz.OpPermission
	var total int64
	switch req.FilterByTarget {
	case dmsV1.OpPermissionTargetAll:
		ops, total, err = d.OpPermissionUsecase.ListOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	case dmsV1.OpPermissionTargetUser:
		ops, total, err = d.OpPermissionUsecase.ListUserOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	case dmsV1.OpPermissionTargetMember:
		ops, total, err = d.OpPermissionUsecase.ListMemberOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid filter by target: %v", req.FilterByTarget)
	}

	ret := make([]*dmsV1.ListOpPermission, len(ops))
	for i, o := range ops {
		opRangeTyp, err := dmsV1.ParseOpRangeType(o.RangeType.String())
		if err != nil {
			return nil, fmt.Errorf("parse op range type failed: %v", err)
		}
		ret[i] = &dmsV1.ListOpPermission{
			OpPermission: dmsV1.UidWithName{Uid: o.GetUID(), Name: o.Name},
			Description:  o.Desc,
			RangeType:    opRangeTyp,
		}
	}

	return &dmsV1.ListOpPermissionReply{
		Payload: struct {
			OpPermissions []*dmsV1.ListOpPermission `json:"op_permissions"`
			Total         int64                     `json:"total"`
		}{OpPermissions: ret, Total: total},
	}, nil
}
