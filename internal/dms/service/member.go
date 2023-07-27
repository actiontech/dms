package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) AddMember(ctx context.Context, currentUserUid string, req *dmsV1.AddMemberReq) (reply *dmsV1.AddMemberReply, err error) {
	d.log.Infof("AddMembers.req=%v", req)
	defer func() {
		d.log.Infof("AddMembers.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	roles := make([]biz.MemberRoleWithOpRange, 0, len(req.Member.RoleWithOpRanges))
	for _, p := range req.Member.RoleWithOpRanges {
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

	uid, err := d.MemberUsecase.CreateMember(ctx, currentUserUid, req.Member.UserUid, req.Member.NamespaceUid, req.Member.IsNamespaceAdmin, roles)
	if err != nil {
		return nil, fmt.Errorf("create member failed: %w", err)
	}

	return &dmsV1.AddMemberReply{
		Payload: struct {
			// member UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) ListMembers(ctx context.Context, req *dmsV1.ListMemberReq) (reply *dmsV1.ListMemberReply, err error) {
	d.log.Infof("ListMembers.req=%v", req)
	defer func() {
		d.log.Infof("ListMembers.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.MemberField
	switch req.OrderBy {
	case dmsV1.MemberOrderByUserUid:
		orderBy = biz.MemberFieldUserUID
	default:
		orderBy = biz.MemberFieldUserUID
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByUserUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.MemberFieldUserUID),
			Operator: pkgConst.FilterOperatorContains,
			Value:    req.FilterByUserUid,
		})
	}
	if req.NamespaceUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.MemberFieldNamespaceUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.NamespaceUid,
		})
	}

	listOption := &biz.ListMembersOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	members, total, err := d.MemberUsecase.ListMember(ctx, listOption, req.NamespaceUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListMember, len(members))
	for i, m := range members {
		user, err := d.UserUsecase.GetUser(ctx, m.UserUID)
		if err != nil {
			return nil, fmt.Errorf("get user failed: %v", err)
		}
		ret[i] = &dmsV1.ListMember{
			MemberUid: m.GetUID(),
			User:      dmsV1.UidWithName{Uid: user.GetUID(), Name: user.Name},
		}

		// 遍历成员的角色&权限范围用于展示
		for _, r := range m.RoleWithOpRanges {
			// 获取角色
			role, err := d.RoleUsecase.GetRole(ctx, r.RoleUID)
			if err != nil {
				return nil, fmt.Errorf("get role failed: %v", err)
			}

			isAdmin, err := d.MemberUsecase.IsMemberNamespaceAdmin(ctx, m.GetUID())
			if err != nil {
				return nil, fmt.Errorf("check member is namespace admin failed: %v", err)
			}

			// 如果是空间管理员namespace admin，则表示拥有该空间的所有权限
			if isAdmin {
				ret[i].IsNamespaceAdmin = true

				// 如果不是空间管理员namespace admin，则展示具体的权限范围
			} else {
				// 获取权限范围类型
				opRangeTyp, err := dmsV1.ParseOpRangeType(r.OpRangeType.String())
				if err != nil {
					return nil, fmt.Errorf("parse op range type failed: %v", err)
				}

				// 获取权限范围
				rangeUidWithNames := []dmsV1.UidWithName{}
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
						return nil, fmt.Errorf("member currently only support the db service op range type, but got type: %v", r.OpRangeType)
					default:
						return nil, fmt.Errorf("unsupported op range type: %v", r.OpRangeType)
					}
				}

				ret[i].RoleWithOpRanges = append(ret[i].RoleWithOpRanges, dmsV1.ListMemberRoleWithOpRange{
					RoleUID:     dmsV1.UidWithName{Uid: role.GetUID(), Name: role.Name},
					OpRangeType: opRangeTyp,
					RangeUIDs:   rangeUidWithNames,
				})
			}
		}
	}

	return &dmsV1.ListMemberReply{
		Payload: struct {
			Members []*dmsV1.ListMember `json:"members"`
			Total   int64               `json:"total"`
		}{Members: ret, Total: total},
	}, nil
}

func (d *DMSService) UpdateMember(ctx context.Context, currentUserUid string, req *dmsV1.UpdateMemberReq) (err error) {
	d.log.Infof("UpdateMember.req=%v", req)
	defer func() {
		d.log.Infof("UpdateMember.req=%v;error=%v", req, err)
	}()

	roles := make([]biz.MemberRoleWithOpRange, 0, len(req.Member.RoleWithOpRanges))
	for _, r := range req.Member.RoleWithOpRanges {

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

	if err = d.MemberUsecase.UpdateMember(ctx, currentUserUid, req.MemberUid, pkgConst.UIDOfNamespaceDefault, /*暂时只支持默认namespace*/
		req.Member.IsNamespaceAdmin, roles); nil != err {
		return fmt.Errorf("update member failed: %v", err)
	}

	return nil
}

func (d *DMSService) DelMember(ctx context.Context, currentUserUid string, req *dmsV1.DelMemberReq) (err error) {
	d.log.Infof("DelMember.req=%v", req)
	defer func() {
		d.log.Infof("DelMember.req=%v;error=%v", req, err)
	}()

	if err := d.MemberUsecase.DelMember(ctx, currentUserUid, req.MemberUid); err != nil {
		return fmt.Errorf("delete member failed: %v", err)
	}

	return nil
}

func (d *DMSService) ListMembersForInternal(ctx context.Context, req *dmsCommonV1.ListMembersForInternalReq) (reply *dmsCommonV1.ListMembersForInternalReply, err error) {
	d.log.Infof("ListMembersForInternal.req=%v", req)
	defer func() {
		d.log.Infof("ListMembersForInternal.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	listOption := &biz.ListMembersOpPermissionOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
	}

	members, total, err := d.OpPermissionVerifyUsecase.ListMembersOpPermissionInNamespace(ctx, req.NamespaceUid, listOption)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsCommonV1.ListMembersForInternalItem, len(members))
	for i, m := range members {

		var opPermission []dmsCommonV1.OpPermissionItem

		for _, op := range m.OpPermissions {
			opTyp, err := convertBizOpPermission(op.OpPermissionUID)
			if err != nil {
				return nil, fmt.Errorf("get user op permission error: %v", err)
			}
			dmsCommonOpTyp, err := dmsCommonV1.ParseOpPermissionType(string(opTyp))
			if err != nil {
				return nil, fmt.Errorf("get dms common user op permission error: %v", err)
			}

			rangeTyp, err := convertBizOpRangeType(op.OpRangeType)
			if err != nil {
				return nil, fmt.Errorf("get user op range type error: %v", err)
			}
			dmsCommonRangeTyp, err := dmsCommonV1.ParseOpRangeType(string(rangeTyp))
			if err != nil {
				return nil, fmt.Errorf("get dms common user op range type error: %v", err)
			}

			opPermission = append(opPermission, dmsCommonV1.OpPermissionItem{
				OpPermissionType: dmsCommonOpTyp,
				RangeType:        dmsCommonRangeTyp,
				RangeUids:        op.RangeUIDs,
			})
		}

		isAdmin, err := d.OpPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, m.UserUid, req.NamespaceUid)
		if err != nil {
			return nil, fmt.Errorf("check user namespace admin error: %v", err)
		}

		ret[i] = &dmsCommonV1.ListMembersForInternalItem{
			MemberUid:              m.MemberUid,
			User:                   dmsCommonV1.UidWithName{Uid: m.UserUid, Name: m.UserName},
			IsAdmin:                isAdmin,
			MemberOpPermissionList: opPermission,
		}
	}

	return &dmsCommonV1.ListMembersForInternalReply{
		Payload: struct {
			Members []*dmsCommonV1.ListMembersForInternalItem `json:"members"`
			Total   int64                                     `json:"total"`
		}{Members: ret, Total: total},
	}, nil
}
