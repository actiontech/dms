package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *DMSService) AddRole(ctx context.Context, currentUserUid string, req *dmsV1.AddRoleReq) (reply *dmsV1.AddRoleReply, err error) {
	d.log.Infof("AddRoles.req=%v", req)
	defer func() {
		d.log.Infof("AddRoles.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	uid, err := d.RoleUsecase.CreateRole(ctx, currentUserUid, req.Role.Name, req.Role.Desc, req.Role.OpPermissionUids)
	if err != nil {
		return nil, fmt.Errorf("create role failed: %w", err)
	}

	return &dmsV1.AddRoleReply{
		Payload: struct {
			// role UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) UpdateRole(ctx context.Context, currentUserUid string, req *dmsV1.UpdateRoleReq) (err error) {
	d.log.Infof("UpdateRole.req=%v", req)
	defer func() {
		d.log.Infof("UpdateRole.req=%v;error=%v", req, err)
	}()

	if err = d.RoleUsecase.UpdateRole(ctx, currentUserUid, req.RoleUid, *req.Role.IsDisabled,
		req.Role.Desc, *req.Role.OpPermissionUids); nil != err {
		return fmt.Errorf("update role failed: %v", err)
	}

	return nil
}

func (d *DMSService) DelRole(ctx context.Context, currentUserUid string, req *dmsV1.DelRoleReq) (err error) {
	d.log.Infof("DelRole.req=%v", req)
	defer func() {
		d.log.Infof("DelRole.req=%v;error=%v", req, err)
	}()

	if err := d.RoleUsecase.DelRole(ctx, currentUserUid, req.RoleUid); err != nil {
		return fmt.Errorf("delete role failed: %v", err)
	}

	return nil
}

func (d *DMSService) ListRoles(ctx context.Context, req *dmsV1.ListRoleReq) (reply *dmsV1.ListRoleReply, err error) {
	d.log.Infof("ListRoles.req=%v", req)
	defer func() {
		d.log.Infof("ListRoles.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.RoleField
	switch req.OrderBy {
	case dmsV1.RoleOrderByName:
		orderBy = biz.RoleFieldName
	default:
		orderBy = biz.RoleFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.RoleFieldName),
			Operator: pkgConst.FilterOperatorContains,
			Value:    req.FilterByName,
		})
	}

	listOption := &biz.ListRolesOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	roles, total, err := d.RoleUsecase.ListRole(ctx, listOption)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListRole, len(roles))
	for i, r := range roles {
		ret[i] = &dmsV1.ListRole{
			RoleUid: r.GetUID(),
			Name:    r.Name,
			Desc:    r.Desc,
		}

		// 获取角色状态
		switch r.Stat {
		case biz.RoleStatOK:
			ret[i].Stat = dmsV1.StatOK
		case biz.RoleStatDisable:
			ret[i].Stat = dmsV1.StatDisable
		default:
			ret[i].Stat = dmsV1.StatUnknown
		}

		// 获取角色的操作权限
		ops, err := d.RoleUsecase.GetOpPermissions(ctx, r.GetUID())
		if err != nil {
			return nil, err
		}
		for _, op := range ops {
			ret[i].OpPermissions = append(ret[i].OpPermissions, dmsV1.UidWithName{Uid: op.GetUID(), Name: op.Name})
		}

	}

	return &dmsV1.ListRoleReply{
		Payload: struct {
			Roles []*dmsV1.ListRole `json:"roles"`
			Total int64             `json:"total"`
		}{Roles: ret, Total: total},
	}, nil
}
