package service

import (
	"context"
	"fmt"
	"github.com/actiontech/dms/internal/pkg/locale"

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

	uid, err := d.MemberUsecase.CreateMember(ctx, currentUserUid, req.Member.UserUid, req.ProjectUid, req.Member.IsProjectAdmin, roles, req.Member.ProjectManagePermissions)
	if err != nil {
		return nil, fmt.Errorf("create member failed: %w", err)
	}

	return &dmsV1.AddMemberReply{
		Data: struct {
			// member UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) ListMemberTips(ctx context.Context, projectId string) (reply *dmsV1.ListMemberTipsReply, err error) {
	members, err := d.OpPermissionVerifyUsecase.ListUsersInProject(ctx, projectId)
	if nil != err {
		return nil, err
	}

	ret := make([]dmsV1.ListMemberTipsItem, len(members))
	for i, m := range members {

		ret[i] = dmsV1.ListMemberTipsItem{
			UserId:   m.UserUid,
			UserName: m.UserName,
		}
	}

	return &dmsV1.ListMemberTipsReply{
		Data: ret,
	}, nil
}

func (d *DMSService) ListMembers(ctx context.Context, req *dmsV1.ListMemberReq) (reply *dmsV1.ListMemberReply, err error) {
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
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByUserUid,
		})
	}
	if req.ProjectUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.MemberFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	listOption := &biz.ListMembersOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	members, total, err := d.MemberUsecase.ListMember(ctx, listOption, req.ProjectUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListMember, len(members))
	for i, m := range members {
		user, err := d.UserUsecase.GetUser(ctx, m.UserUID)
		if err != nil {
			return nil, fmt.Errorf("get user failed: %v", err)
		}
		isGroupMember := false
		memberGroups, err := d.MemberGroupUsecase.GetMemberGroupsByUserIDAndProjectID(ctx, m.UserUID, req.ProjectUid)
		if err != nil {
			return nil, fmt.Errorf("get member groups failed: %v", err)
		}
		if len(memberGroups) > 0 {
			isGroupMember = true
		}

		roleWithOpRanges, err := d.buildRoleWithOpRanges(ctx, m.RoleWithOpRanges, nil)
		if err != nil {
			return nil, err
		}
		projectManagePermissions := make([]dmsV1.ProjectManagePermission, 0,len(m.OpPermissions))
		memberGroupRoleWithOpRanges := make([]dmsV1.ListMemberRoleWithOpRange, 0)
		// 转换所有用户组的RoleWithOpRanges
		for _, memberGroup := range memberGroups {
			memberRoleWithOpRanges, err := d.buildRoleWithOpRanges(ctx, memberGroup.RoleWithOpRanges, memberGroup)
			if err != nil {
				return nil, err
			}
			memberGroupRoleWithOpRanges = append(memberGroupRoleWithOpRanges, memberRoleWithOpRanges...)
			for _, permission := range memberGroup.OpPermissions {
				projectManagePermissions = append(projectManagePermissions, dmsV1.ProjectManagePermission{
					Uid:  permission.GetUID(),
					Name: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionNameByUID[permission.GetUID()]),
					MemberGroup: memberGroup.Name,
				})
			}
		}

		projectOpPermissions := d.aggregateRoleByDataSource(roleWithOpRanges, memberGroupRoleWithOpRanges)
		for _, permission := range m.OpPermissions {
			projectManagePermissions = append(projectManagePermissions, dmsV1.ProjectManagePermission{
				Uid:  permission.GetUID(),
				Name: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionNameByUID[permission.GetUID()]),
			})
		}
		ret[i] = &dmsV1.ListMember{
			MemberUid:        m.GetUID(),
			User:             dmsV1.UidWithName{Uid: user.GetUID(), Name: user.Name},
			RoleWithOpRanges: roleWithOpRanges,
			Projects: 		  m.Projects,
			IsGroupMember:    isGroupMember,
			CurrentProjectOpPermissions: projectOpPermissions,
			CurrentProjectAdmin: d.buildMemberCurrentProjectAdmin(m, memberGroups),
			CurrentProjectManagePermissions: projectManagePermissions,
		}

		for _, r := range m.RoleWithOpRanges {
			if r.RoleUID == pkgConst.UIDOfRoleProjectAdmin {
				ret[i].IsProjectAdmin = true
			}
		}

		// 获取用户的权限
		ops, err := d.UserUsecase.GetUserOpPermissions(ctx, m.UserUID)
		if err != nil {
			return nil, err
		}
		for _, op := range ops {
			if op.RangeType == biz.OpRangeTypeGlobal {
				ret[i].PlatformRoles = append(ret[i].PlatformRoles, dmsV1.UidWithName{
					Uid:  op.GetUID(),
					Name: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionNameByUID[op.GetUID()]),
				})
			}
		}
	}
	return &dmsV1.ListMemberReply{
			Data:  ret,
			Total: total,
		},
		nil
}

func (d *DMSService) buildMemberCurrentProjectAdmin(member *biz.Member, memberGroups []*biz.MemberGroup) dmsV1.CurrentProjectAdmin {
	isProjectAdmin := false
	for _, r := range member.RoleWithOpRanges {
		if r.RoleUID == pkgConst.UIDOfRoleProjectAdmin {
			isProjectAdmin = true
		}
	}
	memberGroupNames := make([]string, 0, len(memberGroups))
	for _, group := range memberGroups {
		memberGroupNames = append(memberGroupNames, group.Name)
		for _, r := range group.RoleWithOpRanges {
			if r.RoleUID == pkgConst.UIDOfRoleProjectAdmin {
				isProjectAdmin = true
			}
		}
	}
	return dmsV1.CurrentProjectAdmin{
		IsAdmin: isProjectAdmin,
		MemberGroups: memberGroupNames,
	}
}

func (d *DMSService) aggregateRoleByDataSource(roleWithOpRanges []dmsV1.ListMemberRoleWithOpRange, memberGroupRoleWithOpRanges []dmsV1.ListMemberRoleWithOpRange) []dmsV1.ProjectOpPermission {
	dataSourceRolesMap := make(map[string][]dmsV1.ProjectRole)
	for _, role := range roleWithOpRanges {
		for _, dataSource := range role.RangeUIDs {
			projectRole := dmsV1.ProjectRole{
				Uid: role.RoleUID.Uid,
				Name: role.RoleUID.Name,
				OpPermissions: role.OpPermissions,
			}
			if _, exists := dataSourceRolesMap[dataSource.Name]; exists {
				dataSourceRolesMap[dataSource.Name] = append(dataSourceRolesMap[dataSource.Name], projectRole)
			} else {
				dataSourceRolesMap[dataSource.Name] = []dmsV1.ProjectRole{projectRole}
			}
		}
	}

	for _, groupRole := range memberGroupRoleWithOpRanges {
		projectRole := dmsV1.ProjectRole{
			Uid: groupRole.RoleUID.Uid,
			Name: groupRole.RoleUID.Name,
			OpPermissions: groupRole.OpPermissions,
			MemberGroup: groupRole.MemberGroup,
		}
		for _, dataSource := range groupRole.RangeUIDs {
			if _, exists := dataSourceRolesMap[dataSource.Name]; exists {
				dataSourceRolesMap[dataSource.Name] = append(dataSourceRolesMap[dataSource.Name], projectRole)
			} else {
				dataSourceRolesMap[dataSource.Name] = []dmsV1.ProjectRole{projectRole}
			}
		}
	}
	projectOpPermissions := make([]dmsV1.ProjectOpPermission, 0, len(dataSourceRolesMap))
	for dataSource, roles := range dataSourceRolesMap {
		projectOpPermissions = append(projectOpPermissions, dmsV1.ProjectOpPermission{DataSource: dataSource, Roles: roles})
	}
	return projectOpPermissions
}

func convert2ProjectMemberGroup(memberGroup *biz.MemberGroup, memberGroupOpPermissions []dmsV1.UidWithName) *dmsV1.ProjectMemberGroup {
	if memberGroup == nil {
		return nil
	}
	users := make([]dmsV1.UidWithName, 0, len(memberGroup.Users))
	for _, user := range memberGroup.Users {
		users = append(users, dmsV1.UidWithName{Uid: user.Uid, Name: user.Name})
	}
	return &dmsV1.ProjectMemberGroup{
		Uid: memberGroup.UID,
		Name: memberGroup.Name,
		Users: users,
		OpPermissions: memberGroupOpPermissions,
	}
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

	if err = d.MemberUsecase.UpdateMember(ctx, currentUserUid, req.MemberUid, pkgConst.UIDOfProjectDefault, /*暂时只支持默认project*/
		req.Member.IsProjectAdmin, roles, req.Member.ProjectManagePermissions); nil != err {
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
	listOption := &biz.ListMembersOpPermissionOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
	}

	members, total, err := d.OpPermissionVerifyUsecase.ListUsersOpPermissionInProject(ctx, req.ProjectUid, listOption)
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

		isAdmin, err := d.OpPermissionVerifyUsecase.IsUserProjectAdmin(ctx, m.UserUid, req.ProjectUid)
		if err != nil {
			return nil, fmt.Errorf("check user project admin error: %v", err)
		}

		ret[i] = &dmsCommonV1.ListMembersForInternalItem{
			User:                   dmsCommonV1.UidWithName{Uid: m.UserUid, Name: m.UserName},
			IsAdmin:                isAdmin,
			MemberOpPermissionList: opPermission,
		}
	}

	return &dmsCommonV1.ListMembersForInternalReply{
		Data: ret, Total: total,
	}, nil
}
