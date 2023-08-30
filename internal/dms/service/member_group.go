package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *DMSService) ListMemberGroups(ctx context.Context, req *dmsV1.ListMemberGroupsReq) (reply *dmsV1.ListMemberGroupsReply, err error) {
	var orderBy biz.MemberGroupField
	switch req.OrderBy {
	case dmsV1.MemberGroupOrderByName:
		orderBy = biz.MemberGroupFieldName
	default:
		orderBy = biz.MemberGroupFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.MemberGroupFieldName),
			Operator: pkgConst.FilterOperatorContains,
			Value:    req.FilterByName,
		})
	}
	if req.NamespaceUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.MemberGroupFieldNamespaceUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.NamespaceUid,
		})
	}

	listOption := &biz.ListMemberGroupsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	memberGroups, total, err := d.MemberGroupUsecase.ListMemberGroups(ctx, listOption, req.NamespaceUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListMemberGroup, 0, len(memberGroups))
	for _, memberGroup := range memberGroups {
		isAdmin, err := d.MemberGroupUsecase.IsMemberGroupNamespaceAdmin(ctx, memberGroup.UID)
		if err != nil {
			return nil, fmt.Errorf("check member group is namespace admin failed: %v", err)
		}

		users := make([]dmsV1.UidWithName, 0, len(memberGroup.Users))
		for _, user := range memberGroup.Users {
			users = append(users, dmsV1.UidWithName{
				Uid:  user.Uid,
				Name: user.Name,
			})
		}

		roleWithOpRanges, err := d.buildMemberGroupRoleWithOpRanges(ctx, memberGroup.RoleWithOpRanges)
		if err != nil {
			return nil, err
		}

		item := &dmsV1.ListMemberGroup{
			Name:             memberGroup.Name,
			Uid:              memberGroup.UID,
			IsNamespaceAdmin: isAdmin,
			Users:            users,
			RoleWithOpRanges: roleWithOpRanges,
		}

		ret = append(ret, item)
	}

	return &dmsV1.ListMemberGroupsReply{
		Payload: struct {
			MemberGroups []*dmsV1.ListMemberGroup `json:"member_groups"`
			Total        int64                    `json:"total"`
		}{MemberGroups: ret, Total: total},
	}, nil
}

func (d *DMSService) buildMemberGroupRoleWithOpRanges(ctx context.Context, roleWithOpRanges []biz.MemberRoleWithOpRange) ([]dmsV1.ListMemberRoleWithOpRange, error) {
	ret := make([]dmsV1.ListMemberRoleWithOpRange, 0, len(roleWithOpRanges))

	// 遍历成员的角色&权限范围用于展示
	for _, r := range roleWithOpRanges {
		if r.RoleUID == pkgConst.UIDOfRoleNamespaceAdmin {
			continue
		}

		// 获取角色
		role, err := d.RoleUsecase.GetRole(ctx, r.RoleUID)
		if err != nil {
			return nil, fmt.Errorf("get role failed: %v", err)
		}

		// 获取权限范围类型
		opRangeTyp, err := dmsV1.ParseOpRangeType(r.OpRangeType.String())
		if err != nil {
			return nil, fmt.Errorf("parse op range type failed: %v", err)
		}

		// 获取权限范围
		rangeUidWithNames := make([]dmsV1.UidWithName, 0)
		for _, uid := range r.RangeUIDs {
			switch r.OpRangeType {
			case biz.OpRangeTypeDBService:
				dbService, err := d.DBServiceUsecase.GetDBService(ctx, uid)
				if err != nil {
					return nil, fmt.Errorf("get db service failed: %v", err)
				}
				rangeUidWithNames = append(rangeUidWithNames, dmsV1.UidWithName{Uid: dbService.GetUID(), Name: dbService.Name})
			// 成员目前只支持配置数据源范围的权限
			case biz.OpRangeTypeNamespace, biz.OpRangeTypeGlobal:
				//return nil, fmt.Errorf("member currently only support the db service op range type, but got type: %v", r.OpRangeType)
			default:
				return nil, fmt.Errorf("unsupported op range type: %v", r.OpRangeType)
			}
		}

		ret = append(ret, dmsV1.ListMemberRoleWithOpRange{
			RoleUID:     dmsV1.UidWithName{Uid: role.GetUID(), Name: role.Name},
			OpRangeType: opRangeTyp,
			RangeUIDs:   rangeUidWithNames,
		})
	}

	return ret, nil
}

func (d *DMSService) GetMemberGroup(ctx context.Context, req *dmsV1.GetMemberGroupReq) (reply *dmsV1.GetMemberGroupReply, err error) {
	memberGroup, err := d.MemberGroupUsecase.GetMemberGroup(ctx, req.MemberGroupUid, req.NamespaceId)
	if err != nil {
		return nil, err
	}

	isAdmin, err := d.MemberGroupUsecase.IsMemberGroupNamespaceAdmin(ctx, memberGroup.UID)
	if err != nil {
		return nil, fmt.Errorf("check member group is namespace admin failed: %v", err)
	}

	users := make([]dmsV1.UidWithName, 0, len(memberGroup.Users))
	for _, user := range memberGroup.Users {
		users = append(users, dmsV1.UidWithName{
			Uid:  user.Uid,
			Name: user.Name,
		})
	}

	roleWithOpRanges, err := d.buildMemberGroupRoleWithOpRanges(ctx, memberGroup.RoleWithOpRanges)
	if err != nil {
		return nil, err
	}

	ret := &dmsV1.GetMemberGroup{
		Name:             memberGroup.Name,
		Uid:              memberGroup.UID,
		IsNamespaceAdmin: isAdmin,
		Users:            users,
		RoleWithOpRanges: roleWithOpRanges,
	}

	if err != nil {
		return nil, err
	}

	return &dmsV1.GetMemberGroupReply{
		Payload: struct {
			MemberGroup *dmsV1.GetMemberGroup `json:"member_group"`
		}{ret},
	}, nil
}

func (d *DMSService) AddMemberGroup(ctx context.Context, currentUserUid string, req *dmsV1.AddMemberGroupReq) (reply *dmsV1.AddMemberReply, err error) {
	roles := make([]biz.MemberRoleWithOpRange, 0, len(req.MemberGroup.RoleWithOpRanges))
	for _, p := range req.MemberGroup.RoleWithOpRanges {
		typ, err := biz.ParseOpRangeType(string(p.OpRangeType))
		if err != nil {
			return nil, fmt.Errorf("parse op range type failed: %v", err)
		}
		roles = append(roles, biz.MemberRoleWithOpRange{
			RoleUID:     p.RoleUID,
			OpRangeType: typ,
			RangeUIDs:   p.RangeUIDs,
		})
	}

	params := &biz.MemberGroup{
		IsNamespaceAdmin: req.MemberGroup.IsNamespaceAdmin,
		Name:             req.MemberGroup.Name,
		NamespaceUID:     req.MemberGroup.NamespaceUid,
		UserUids:         req.MemberGroup.UserUids,
		RoleWithOpRanges: roles,
	}

	uid, err := d.MemberGroupUsecase.CreateMemberGroup(ctx, currentUserUid, params)
	if err != nil {
		return nil, fmt.Errorf("create member group failed: %w", err)
	}

	return &dmsV1.AddMemberReply{
		Payload: struct {
			// member UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) UpdateMemberGroup(ctx context.Context, currentUserUid string, req *dmsV1.UpdateMemberGroupReq) (err error) {
	roles := make([]biz.MemberRoleWithOpRange, 0, len(req.MemberGroup.RoleWithOpRanges))
	for _, r := range req.MemberGroup.RoleWithOpRanges {

		typ, err := biz.ParseOpRangeType(string(r.OpRangeType))
		if err != nil {
			return fmt.Errorf("parse op range type failed: %v", err)
		}
		roles = append(roles, biz.MemberRoleWithOpRange{
			RoleUID:     r.RoleUID,
			OpRangeType: typ,
			RangeUIDs:   r.RangeUIDs,
		})
	}

	params := &biz.MemberGroup{
		UID:              req.MemberGroupUid,
		IsNamespaceAdmin: req.MemberGroup.IsNamespaceAdmin,
		NamespaceUID:     req.MemberGroup.NamespaceUid,
		UserUids:         req.MemberGroup.UserUids,
		RoleWithOpRanges: roles,
	}

	err = d.MemberGroupUsecase.UpdateMemberGroup(ctx, currentUserUid, params)
	if err != nil {
		return fmt.Errorf("update member group failed: %w", err)
	}

	return nil
}

func (d *DMSService) DeleteMemberGroup(ctx context.Context, currentUserUid string, req *dmsV1.DeleteMemberGroupReq) (err error) {
	if err = d.MemberGroupUsecase.DeleteMemberGroup(ctx, currentUserUid, req.MemberGroupUid, req.NamespaceId); err != nil {
		return fmt.Errorf("delete member group failed: %v", err)
	}

	return nil
}
