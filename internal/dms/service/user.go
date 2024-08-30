package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	jwtPkg "github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/golang-jwt/jwt/v4"
)

func (d *DMSService) VerifyUserLogin(ctx context.Context, req *dmsV1.VerifyUserLoginReq) (reply *dmsV1.VerifyUserLoginReply, err error) {
	d.log.Infof("VerifyUserLogin.req=%v", req)
	defer func() {
		d.log.Infof("VerifyUserLogin.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var verifyFailedMsg string
	uid, err := d.UserUsecase.UserLogin(ctx, req.UserName, req.Password)
	if nil != err {
		verifyFailedMsg = err.Error()
	}

	return &dmsV1.VerifyUserLoginReply{
		Data: struct {
			// If verify Successful, return empty string, otherwise return error message
			VerifyFailedMsg string `json:"verify_failed_msg"`
			// If verify Successful, return user uid
			UserUid string `json:"user_uid"`
		}{UserUid: uid, VerifyFailedMsg: verifyFailedMsg},
	}, nil
}

func (d *DMSService) AfterUserLogin(ctx context.Context, req *dmsV1.AfterUserLoginReq) (err error) {
	d.log.Infof("AfterUserLogin.req=%v", req)
	defer func() {
		d.log.Infof("AfterUserLogin.req=%v;error=%v", req, err)
	}()

	err = d.UserUsecase.AfterUserLogin(ctx, req.UserUid)
	if nil != err {
		return fmt.Errorf("handle after user login error: %v", err)
	}
	return nil
}

func (d *DMSService) GetCurrentUser(ctx context.Context, req *dmsV1.GetUserBySessionReq) (reply *dmsV1.GetUserBySessionReply, err error) {
	d.log.Infof("GetCurrentUser,req=%v", req)
	defer func() {
		d.log.Infof("GetCurrentUser.req=%v,reply=%v;error=%v", req, reply, err)
	}()

	user, err := d.UserUsecase.GetUser(ctx, req.UserUid)
	if nil != err {
		return nil, err
	}

	return &dmsV1.GetUserBySessionReply{
		Data: struct {
			// User UID
			UserUid string `json:"user_uid"`
			// User name
			Name string `json:"name"`
		}{UserUid: user.GetUID(), Name: user.Name},
	}, nil
}

func (d *DMSService) AddUser(ctx context.Context, currentUserUid string, req *dmsV1.AddUserReq) (reply *dmsV1.AddUserReply, err error) {
	d.log.Infof("AddUsers.req=%v", req)
	defer func() {
		d.log.Infof("AddUsers.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	args := &biz.CreateUserArgs{
		Name:             req.User.Name,
		Desc:             req.User.Desc,
		Password:         req.User.Password,
		Email:            req.User.Email,
		Phone:            req.User.Phone,
		WxID:             req.User.WxID,
		IsDisabled:       false,
		UserGroupUIDs:    req.User.UserGroupUids,
		OpPermissionUIDs: req.User.OpPermissionUids,
	}

	uid, err := d.UserUsecase.CreateUser(ctx, currentUserUid, args)
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return &dmsV1.AddUserReply{
		Data: struct {
			// user UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) UpdateUser(ctx context.Context, req *dmsV1.UpdateUserReq, currentUserUid string) (err error) {
	/// TODO 当前保留了用户组概念，但暂时未有用户组实际应用场景.前端移除用户组相关功能，默认关联用户组为空
	if req.User.UserGroupUids == nil {
		req.User.UserGroupUids = &[]string{}
	}

	if err = d.UserUsecase.UpdateUser(ctx, currentUserUid, req.UserUid, *req.User.IsDisabled,
		req.User.Password, req.User.Email, req.User.Phone, req.User.WxID, *req.User.UserGroupUids, *req.User.OpPermissionUids); nil != err {
		return fmt.Errorf("update user failed: %v", err)
	}

	return nil
}

func (d *DMSService) UpdateCurrentUser(ctx context.Context, req *dmsV1.UpdateCurrentUserReq, currentUserUid string) (err error) {
	if err = d.UserUsecase.UpdateCurrentUser(ctx, currentUserUid, req.User.OldPassword, req.User.Password, req.User.Email, req.User.Phone, req.User.WxID); nil != err {
		return fmt.Errorf("update user failed: %v", err)
	}

	return nil
}

func (d *DMSService) DelUser(ctx context.Context, currentUserUid string, req *dmsV1.DelUserReq) (err error) {
	d.log.Infof("DelUser.req=%v", req)
	defer func() {
		d.log.Infof("DelUser.req=%v;error=%v", req, err)
	}()

	if err := d.UserUsecase.DelUser(ctx, currentUserUid, req.UserUid); err != nil {
		return fmt.Errorf("delete user failed: %v", err)
	}

	return nil
}

func (d *DMSService) ListUsers(ctx context.Context, req *dmsCommonV1.ListUserReq) (reply *dmsCommonV1.ListUserReply, err error) {
	d.log.Infof("ListUsers.req=%v", req)
	defer func() {
		d.log.Infof("ListUsers.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.UserField
	switch req.OrderBy {
	case dmsCommonV1.UserOrderByName:
		orderBy = biz.UserFieldName
	default:
		orderBy = biz.UserFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.UserFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}

	if req.FilterByUids != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.UserFieldUID),
			Operator: pkgConst.FilterOperatorIn,
			Value:    strings.Split(req.FilterByUids, ","),
		})
	}

	// 默认为false,不展示已删除用户
	if !req.FilterDeletedUser {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.UserFieldDeletedAt),
			Operator: pkgConst.FilterOperatorIsNull,
			Value:    nil,
		})
	}

	listOption := &biz.ListUsersOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	users, total, err := d.UserUsecase.ListUser(ctx, listOption)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsCommonV1.ListUser, len(users))
	for i, u := range users {
		ret[i] = &dmsCommonV1.ListUser{
			UserUid:            u.GetUID(),
			Name:               u.Name,
			Email:              u.Email,
			Phone:              u.Phone,
			WxID:               u.WxID,
			IsDeleted:          u.Deleted,
			ThirdPartyUserInfo: u.ThirdPartyUserInfo,
		}
		// 已删除用户只有基础信息
		if u.Deleted {
			continue
		}
		// 获取用户状态
		switch u.Stat {
		case biz.UserStatOK:
			ret[i].Stat = dmsCommonV1.StatOK
		case biz.UserStatDisable:
			ret[i].Stat = dmsCommonV1.StatDisable
		default:
			ret[i].Stat = dmsCommonV1.StatUnknown
		}

		// 获取用户鉴权类型
		switch u.UserAuthenticationType {
		case biz.UserAuthenticationTypeDMS:
			ret[i].AuthenticationType = dmsCommonV1.UserAuthenticationTypeDMS
		case biz.UserAuthenticationTypeLDAP:
			ret[i].AuthenticationType = dmsCommonV1.UserAuthenticationTypeLDAP
		case biz.UserAuthenticationTypeOAUTH2:
			ret[i].AuthenticationType = dmsCommonV1.UserAuthenticationTypeOAUTH2
		default:
			ret[i].AuthenticationType = dmsCommonV1.UserAuthenticationTypeUnknown
		}

		// 获取用户所属的用户组
		groups, err := d.UserUsecase.GetUserGroups(ctx, u.GetUID())
		if err != nil {
			return nil, err
		}
		for _, g := range groups {
			ret[i].UserGroups = append(ret[i].UserGroups, dmsCommonV1.UidWithName{Uid: g.GetUID(), Name: g.Name})
		}

		// 获取用户的权限
		ops, err := d.UserUsecase.GetUserOpPermissions(ctx, u.GetUID())
		if err != nil {
			return nil, err
		}
		for _, op := range ops {
			ret[i].OpPermissions = append(ret[i].OpPermissions, dmsCommonV1.UidWithName{Uid: op.GetUID(), Name: op.Name})
		}

	}

	return &dmsCommonV1.ListUserReply{
		Data: ret, Total: total,
	}, nil
}

func (d *DMSService) AddUserGroup(ctx context.Context, currentUserUid string, req *dmsV1.AddUserGroupReq) (reply *dmsV1.AddUserGroupReply, err error) {
	d.log.Infof("AddUserGroups.req=%v", req)
	defer func() {
		d.log.Infof("AddUserGroups.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	args := &biz.CreateUserGroupArgs{
		Name:     req.UserGroup.Name,
		Desc:     req.UserGroup.Desc,
		UserUids: req.UserGroup.UserUids,
	}

	uid, err := d.UserGroupUsecase.CreateUserGroup(ctx, currentUserUid, args)
	if err != nil {
		return nil, fmt.Errorf("create user group failed: %w", err)
	}

	return &dmsV1.AddUserGroupReply{
		Data: struct {
			// user group UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) UpdateUserGroup(ctx context.Context, currentUserUid string, req *dmsV1.UpdateUserGroupReq) (err error) {
	d.log.Infof("UpdateUserGroup.req=%v", req)
	defer func() {
		d.log.Infof("UpdateUserGroup.req=%v;error=%v", req, err)
	}()

	if err = d.UserGroupUsecase.UpdateUserGroup(ctx, currentUserUid, req.UserGroupUid, *req.UserGroup.IsDisabled,
		req.UserGroup.Desc, *req.UserGroup.UserUids); nil != err {
		return fmt.Errorf("update user group failed: %v", err)
	}

	return nil
}

func (d *DMSService) DelUserGroup(ctx context.Context, currentUserUid string, req *dmsV1.DelUserGroupReq) (err error) {
	d.log.Infof("DelUserGroup.req=%v", req)
	defer func() {
		d.log.Infof("DelUserGroup.req=%v;error=%v", req, err)
	}()

	if err := d.UserGroupUsecase.DelUserGroup(ctx, currentUserUid, req.UserGroupUid); err != nil {
		return fmt.Errorf("delete user group failed: %v", err)
	}

	return nil
}

func (d *DMSService) ListUserGroups(ctx context.Context, req *dmsV1.ListUserGroupReq) (reply *dmsV1.ListUserGroupReply, err error) {
	d.log.Infof("ListUserGroups.req=%v", req)
	defer func() {
		d.log.Infof("ListUserGroups.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.UserGroupField
	switch req.OrderBy {
	case dmsV1.UserGroupOrderByName:
		orderBy = biz.UserGroupFieldName
	default:
		orderBy = biz.UserGroupFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.UserGroupFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}

	listOption := &biz.ListUserGroupsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	groups, total, err := d.UserGroupUsecase.ListUserGroup(ctx, listOption)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListUserGroup, len(groups))
	for i, g := range groups {
		ret[i] = &dmsV1.ListUserGroup{
			UserGroupUid: g.GetUID(),
			Name:         g.Name,
			Desc:         g.Desc,
		}

		// 获取用户组状态
		switch g.Stat {
		case biz.UserGroupStatOK:
			ret[i].Stat = dmsV1.StatOK
		case biz.UserGroupStatDisable:
			ret[i].Stat = dmsV1.StatDisable
		default:
			ret[i].Stat = dmsV1.StatUnknown
		}

		// 获取用户所属的用户组
		users, err := d.UserGroupUsecase.GetUsersInUserGroup(ctx, g.GetUID())
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			ret[i].Users = append(ret[i].Users, dmsV1.UidWithName{Uid: u.GetUID(), Name: u.Name})
		}

	}

	return &dmsV1.ListUserGroupReply{
		Data: ret, Total: total,
	}, nil
}

func (d *DMSService) GetUserOpPermission(ctx context.Context, req *dmsCommonV1.GetUserOpPermissionReq) (reply *dmsCommonV1.GetUserOpPermissionReply, err error) {
	// 兼容新旧版本获取项目ID方式
	projectUid := req.ProjectUid
	if projectUid == "" && req.UserOpPermission != nil {
		projectUid = req.UserOpPermission.ProjectUid
	}

	isAdmin, err := d.OpPermissionVerifyUsecase.IsUserProjectAdmin(ctx, req.UserUid, projectUid)
	if err != nil {
		return nil, fmt.Errorf("check user admin error: %v", err)
	}

	permissions, err := d.OpPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, req.UserUid, projectUid)
	if err != nil {
		return nil, fmt.Errorf("get user op permission error: %v", err)
	}

	var replyOpPermission = make([]dmsCommonV1.OpPermissionItem, 0, len(permissions))
	for _, p := range permissions {

		opTyp, err := convertBizOpPermission(p.OpPermissionUID)
		if err != nil {
			return nil, fmt.Errorf("get user op permission error: %v", err)
		}
		dmsCommonOpTyp, err := dmsCommonV1.ParseOpPermissionType(string(opTyp))
		if err != nil {
			return nil, fmt.Errorf("get dms common user op permission error: %v", err)
		}

		rangeTyp, err := convertBizOpRangeType(p.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("get user op range type error: %v", err)
		}
		dmsCommonRangeTyp, err := dmsCommonV1.ParseOpRangeType(string(rangeTyp))
		if err != nil {
			return nil, fmt.Errorf("get dms common user op range type error: %v", err)
		}

		replyOpPermission = append(replyOpPermission, dmsCommonV1.OpPermissionItem{
			OpPermissionType: dmsCommonOpTyp,
			RangeType:        dmsCommonRangeTyp,
			RangeUids:        p.RangeUIDs,
		})
	}

	reply = &dmsCommonV1.GetUserOpPermissionReply{
		Data: struct {
			IsAdmin          bool                           `json:"is_admin"`
			OpPermissionList []dmsCommonV1.OpPermissionItem `json:"op_permission_list"`
		}{IsAdmin: isAdmin, OpPermissionList: replyOpPermission},
	}

	d.log.Infof("GetUserOpPermission.resp=%v", reply)
	return reply, nil
}

func (d *DMSService) GetUser(ctx context.Context, req *dmsCommonV1.GetUserReq) (reply *dmsCommonV1.GetUserReply, err error) {
	d.log.Infof("GetUser.req=%v", req)
	defer func() {
		d.log.Infof("GetUser.req=%v;error=%v", req, err)
	}()

	u, err := d.UserUsecase.GetUser(ctx, req.UserUid)
	if err != nil {
		return nil, fmt.Errorf("get user error: %v", err)
	}

	dmsCommonUser := &dmsCommonV1.GetUser{
		UserUid:            u.GetUID(),
		Name:               u.Name,
		Email:              u.Email,
		Phone:              u.Phone,
		WxID:               u.WxID,
		ThirdPartyUserInfo: u.ThirdPartyUserInfo,
	}

	// 获取用户状态
	switch u.Stat {
	case biz.UserStatOK:
		dmsCommonUser.Stat = dmsCommonV1.StatOK
	case biz.UserStatDisable:
		dmsCommonUser.Stat = dmsCommonV1.StatDisable
	default:
		dmsCommonUser.Stat = dmsCommonV1.StatUnknown
	}

	// 获取用户鉴权类型
	switch u.UserAuthenticationType {
	case biz.UserAuthenticationTypeDMS:
		dmsCommonUser.AuthenticationType = dmsCommonV1.UserAuthenticationTypeDMS
	case biz.UserAuthenticationTypeLDAP:
		dmsCommonUser.AuthenticationType = dmsCommonV1.UserAuthenticationTypeLDAP
	case biz.UserAuthenticationTypeOAUTH2:
		dmsCommonUser.AuthenticationType = dmsCommonV1.UserAuthenticationTypeOAUTH2
	default:
		dmsCommonUser.AuthenticationType = dmsCommonV1.UserAuthenticationTypeUnknown
	}

	// 获取用户所属的用户组
	groups, err := d.UserUsecase.GetUserGroups(ctx, u.GetUID())
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		dmsCommonUser.UserGroups = append(dmsCommonUser.UserGroups, dmsCommonV1.UidWithName{Uid: g.GetUID(), Name: g.Name})
	}

	// 获取用户的权限
	ops, err := d.UserUsecase.GetUserOpPermissions(ctx, u.GetUID())
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		dmsCommonUser.OpPermissions = append(dmsCommonUser.OpPermissions, dmsCommonV1.UidWithName{Uid: op.GetUID(), Name: op.Name})
	}
	isAdmin, err := d.UserUsecase.OpPermissionVerifyUsecase.IsUserDMSAdmin(ctx, u.GetUID())
	if err != nil {
		return nil, fmt.Errorf("failed to check user is dms admin")
	}
	dmsCommonUser.IsAdmin = isAdmin
	// 获取管理项目
	userBindProjects := make([]dmsCommonV1.UserBindProject, 0)
	if !isAdmin {
		projectWithOpPermissions, err := d.OpPermissionVerifyUsecase.GetUserProjectOpPermission(ctx, u.GetUID())
		if err != nil {
			return nil, fmt.Errorf("failed to get user project with op permission")
		}
		userBindProjects = d.OpPermissionVerifyUsecase.GetUserManagerProject(ctx, projectWithOpPermissions)
	} else {
		projects, _, err := d.ProjectUsecase.ListProject(ctx, &biz.ListProjectsOption{
			PageNumber:   1,
			LimitPerPage: 999,
		}, u.UID)
		if err != nil {
			return nil, err
		}
		for _, project := range projects {
			userBindProjects = append(userBindProjects, dmsCommonV1.UserBindProject{ProjectID: project.UID, ProjectName: project.Name, IsManager: true})
		}
	}
	dmsCommonUser.UserBindProjects = userBindProjects

	// 获取用户access token
	tokenInfo, err := d.UserUsecase.GetAccessTokenByUser(ctx, u.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user access token: %v", err)
	}
	accessToken := dmsCommonV1.AccessTokenInfo{}
	accessToken.AccessToken = tokenInfo.Token
	accessToken.ExpiredTime = tokenInfo.ExpiredTime.Format("2006-01-02T15:04:05-07:00")
	if tokenInfo.ExpiredTime.Before(time.Now()) {
		accessToken.IsExpired = true
	}
	dmsCommonUser.AccessTokenInfo = accessToken

	reply = &dmsCommonV1.GetUserReply{
		Data: dmsCommonUser,
	}

	d.log.Infof("GetUser.resp=%v", reply)
	return reply, nil
}

func (d *DMSService) GenAccessToken(ctx context.Context, currentUserUid string, req *dmsCommonV1.GenAccessToken) (reply *dmsCommonV1.GenAccessTokenReply, err error) {
	days, err := strconv.ParseUint(req.ExpirationDays, 10, 64)
	if err != nil {
		return nil, err
	}

	expiredTime := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	token, err := jwtPkg.GenJwtTokenWithExpirationTime(jwt.NewNumericDate(expiredTime), jwtPkg.WithUserId(currentUserUid), jwtPkg.WithAccessTokenMark(biz.AccessTokenLogin))
	if err != nil {
		return nil, fmt.Errorf("gen access token failed: %v", err)
	}
	if err := d.UserUsecase.SaveAccessToken(ctx, currentUserUid, token, expiredTime); err != nil {
		return nil, fmt.Errorf("save access token failed: %v", err)
	}

	reply = &dmsCommonV1.GenAccessTokenReply{
		Data: &dmsCommonV1.AccessTokenInfo{
			AccessToken: token,
			ExpiredTime: expiredTime.Format("2006-01-02T15:04:05-07:00"),
		},
	}

	return reply, nil
}

func convertBizOpPermission(opPermissionUid string) (apiOpPermissionTyp dmsCommonV1.OpPermissionType, err error) {
	switch opPermissionUid {
	case pkgConst.UIDOfOpPermissionCreateWorkflow:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeCreateWorkflow
	case pkgConst.UIDOfOpPermissionAuditWorkflow:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeAuditWorkflow
	case pkgConst.UIDOfOpPermissionAuthDBServiceData:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeAuthDBServiceData
	case pkgConst.UIDOfOpPermissionProjectAdmin:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeProjectAdmin
	case pkgConst.UIDOfOpPermissionCreateProject:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeCreateProject
	case pkgConst.UIDOfOpPermissionExecuteWorkflow:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeExecuteWorkflow
	case pkgConst.UIDOfOpPermissionViewOthersWorkflow:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeViewOthersWorkflow
	case pkgConst.UIDOfOpPermissionSaveAuditPlan:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeSaveAuditPlan
	case pkgConst.UIDOfOpPermissionViewOthersAuditPlan:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeViewOtherAuditPlan
	case pkgConst.UIDOfOpPermissionSQLQuery:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeSQLQuery
	case pkgConst.UIDOfOpPermissionExportApprovalReject:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeAuditExportWorkflow
	case pkgConst.UIDOfOpPermissionExportCreate:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeExportCreate
	case pkgConst.UIDOfOpPermissionCreateOptimization:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeCreateOptimization
	case pkgConst.UIDOfOpPermissionViewOthersOptimization:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeViewOthersOptimization
	case pkgConst.UIDOfOpPermissionCreatePipeline:
		apiOpPermissionTyp = dmsCommonV1.OpPermissionTypeCreatePipeline
	default:
		return dmsCommonV1.OpPermissionTypeUnknown, fmt.Errorf("get user op permission type error: invalid op permission uid: %v", opPermissionUid)

	}
	return apiOpPermissionTyp, nil
}

func convertBizOpRangeType(bizOpRangeTyp biz.OpRangeType) (typ dmsV1.OpRangeType, err error) {
	switch bizOpRangeTyp {
	case biz.OpRangeTypeGlobal:
		typ = dmsV1.OpRangeTypeGlobal
	case biz.OpRangeTypeProject:
		typ = dmsV1.OpRangeTypeProject
	case biz.OpRangeTypeDBService:
		typ = dmsV1.OpRangeTypeDBService
	default:
		return dmsV1.OpRangeTypeUnknown, fmt.Errorf("get user op range type error: invalid op range type: %v", bizOpRangeTyp)
	}

	return typ, nil
}

func convertBizUidWithName(bizUidWithNames []biz.UIdWithName) []dmsV1.UidWithName {
	ret := make([]dmsV1.UidWithName, 0)
	for _, v := range bizUidWithNames {
		ret = append(ret, dmsV1.UidWithName{
			Uid:  v.Uid,
			Name: v.Name,
		})
	}
	return ret
}
