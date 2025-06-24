package service

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/pkg/locale"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

var RoleNameByUID = map[string]*i18n.Message{
	pkgConst.UIDOfRoleProjectAdmin: locale.NameRoleProjectAdmin,
	pkgConst.UIDOfRoleDevEngineer:  locale.NameRoleDevEngineer,
	pkgConst.UIDOfRoleDevManager:   locale.NameRoleDevManager,
	pkgConst.UIDOfRoleOpsEngineer:  locale.NameRoleOpsEngineer,
}

var RoleDescByUID = map[string]*i18n.Message{
	pkgConst.UIDOfRoleProjectAdmin: locale.DescRoleProjectAdmin,
	pkgConst.UIDOfRoleDevEngineer:  locale.DescRoleDevEngineer,
	pkgConst.UIDOfRoleDevManager:   locale.DescRoleDevManager,
	pkgConst.UIDOfRoleOpsEngineer:  locale.DescRoleOpsEngineer,
}

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
		Data: struct {
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
		orderBy = biz.RoleFieldCreatedAt
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.RoleFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}

	if req.FuzzyKeyword != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:         string(biz.RoleFieldOpPermission),
			Operator:      pkgConst.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
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
		if r.UID == pkgConst.UIDOfRoleProjectAdmin || r.UID == pkgConst.UIDOfRoleDevEngineer || r.UID == pkgConst.UIDOfRoleDevManager || r.UID == pkgConst.UIDOfRoleOpsEngineer {
			// built in role, localize name and desc
			r.Name = locale.Bundle.LocalizeMsgByCtx(ctx, RoleNameByUID[r.GetUID()])
		}
		ret[i] = &dmsV1.ListRole{
			RoleUid: r.GetUID(),
			Name:    r.Name,
			Desc:    r.Desc,
		}

		// 获取角色状态
		switch r.Stat {
		case biz.RoleStatOK:
			ret[i].Stat = dmsCommonV1.Stat(locale.Bundle.LocalizeMsgByCtx(ctx, locale.StatOK))
		case biz.RoleStatDisable:
			ret[i].Stat = dmsCommonV1.Stat(locale.Bundle.LocalizeMsgByCtx(ctx, locale.StatDisable))
		default:
			ret[i].Stat = dmsCommonV1.Stat(locale.Bundle.LocalizeMsgByCtx(ctx, locale.StatUnknown))
		}

		// 获取角色的操作权限
		ops, err := d.RoleUsecase.GetOpPermissions(ctx, r.GetUID())
		if err != nil {
			return nil, err
		}
		for _, op := range ops {
			// 不支持智能调优时，隐藏相关权限
			if !conf.IsOptimizationEnabled() &&
				(op.UID == pkgConst.UIDOfOpPermissionCreateOptimization || op.UID == pkgConst.UIDOfOpPermissionViewOthersOptimization) {
				continue
			}
			ret[i].OpPermissions = append(ret[i].OpPermissions, dmsV1.ListRoleOpPermission{
				Uid:  op.GetUID(),
				Name: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionNameByUID[op.GetUID()]),
				Module: string(op.Module),
			})
		}

	}

	return &dmsV1.ListRoleReply{
		Data: ret, Total: total,
	}, nil
}
