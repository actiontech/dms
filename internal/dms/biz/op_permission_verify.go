package biz

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OpPermissionVerifyRepo interface {
	IsUserHasOpPermissionInProject(ctx context.Context, userUid, projectUid, opPermissionUid string) (has bool, err error)
	GetUserOpPermissionInProject(ctx context.Context, userUid, projectUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserProjectOpPermissionInProject(ctx context.Context, userUid, projectUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserOpPermission(ctx context.Context, userUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserProjectOpPermission(ctx context.Context, userUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserGlobalOpPermission(ctx context.Context, userUid string) (opPermissions []*OpPermission, err error)
	GetUserProjectWithOpPermissions(ctx context.Context, userUid string) (projectWithPermission []ProjectOpPermissionWithOpRange, err error)
	ListUsersOpPermissionInProject(ctx context.Context, projectUid string, opt *ListMembersOpPermissionOption) (items []ListMembersOpPermissionItem, total int64, err error)
	GetUserProject(ctx context.Context, userUid string) (projects []*Project, err error)
	ListUsersInProject(ctx context.Context, projectUid string) (items []ListMembersOpPermissionItem, err error)
}

type OpPermissionVerifyUsecase struct {
	tx   TransactionGenerator
	repo OpPermissionVerifyRepo
	log  *utilLog.Helper
}

func NewOpPermissionVerifyUsecase(log utilLog.Logger, tx TransactionGenerator, repo OpPermissionVerifyRepo) *OpPermissionVerifyUsecase {
	return &OpPermissionVerifyUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.op_permission_verify")),
	}
}

func (o *OpPermissionVerifyUsecase) IsUserProjectAdmin(ctx context.Context, userUid, projectUid string) (bool, error) {
	// 内置用户admin和sys拥有所有权限
	switch userUid {
	case pkgConst.UIDOfUserAdmin, pkgConst.UIDOfUserSys:
		return true, nil
	default:
	}
	has, err := o.repo.IsUserHasOpPermissionInProject(ctx, userUid, projectUid, pkgConst.UIDOfOpPermissionProjectAdmin)
	if err != nil {
		return false, fmt.Errorf("failed to check user is project admin: %v", err)
	}
	return has, nil
}

func (o *OpPermissionVerifyUsecase) IsUserDMSAdmin(ctx context.Context, userUid string) (bool, error) {
	// 暂且只有内置用户admin和sys拥有平台管理权限
	switch userUid {
	case pkgConst.UIDOfUserAdmin, pkgConst.UIDOfUserSys:
		return true, nil
	default:
		return false, nil
	}
}

func (o *OpPermissionVerifyUsecase) HasGlobalManagementOrViewPermission(ctx context.Context, userUid string) (bool, error) {
	ops, err := o.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return false, err
	}

	for _, op := range ops {
		if op.OpPermissionUID == pkgConst.UIDOfOpPermissionGlobalManagement || op.OpPermissionUID == pkgConst.UIDOfOpPermissionGlobalView {
			return true, nil
		}
	}

	return false, nil
}

func (o *OpPermissionVerifyUsecase) CanOpGlobal(ctx context.Context, userUid string) (bool, error) {
	isUserDMSAdmin, err := o.IsUserDMSAdmin(ctx, userUid)
	if err != nil {
		return false, err
	}
	if isUserDMSAdmin {
		return true, nil
	}

	ops, err := o.repo.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return false, err
	}

	for _, op := range ops {
		if op.UID == pkgConst.UIDOfOpPermissionGlobalManagement {
			return true, nil
		}
	}

	return false, nil
}

func (o *OpPermissionVerifyUsecase) CanOpProject(ctx context.Context, userUid, projectUid string) (bool, error) {
	canGlobalOp, err := o.CanOpGlobal(ctx, userUid)
	if err != nil {
		return false, err
	}
	if canGlobalOp {
		return true, nil
	}

	has, err := o.repo.IsUserHasOpPermissionInProject(ctx, userUid, projectUid, pkgConst.UIDOfOpPermissionProjectAdmin)
	if err != nil {
		return false, fmt.Errorf("failed to check user is project admin: %v", err)
	}

	return has, nil
}

func (o *OpPermissionVerifyUsecase) CanViewProject(ctx context.Context, userUid, projectUid string, uIdOfPermission string) (bool, error) {
	canViewGlobal, err := o.CanViewGlobal(ctx, userUid)
	if err != nil {
		return false, err
	}
	if canViewGlobal {
		return true, nil
	}
	hasPermission := false
	if uIdOfPermission != "" {
		hasPermission, err = o.repo.IsUserHasOpPermissionInProject(ctx, userUid, projectUid, uIdOfPermission)
		if err != nil {
			return false, fmt.Errorf("failed to check user is project admin: %v", err)
		}
	}
	return hasPermission, nil
}

func (o *OpPermissionVerifyUsecase) HasViewPermission(ctx context.Context, userId, projectUid string, uIdOfPermission string) (bool, error) {
	canViewOperationRecord, err := o.HasOpPermissionInProject(ctx, userId, projectUid, uIdOfPermission)
	if err != nil {
		return false, err
	}
	isUserProjectAdmin, err := o.IsUserProjectAdmin(ctx, userId, projectUid)
	if err != nil {
		return false, err
	}
	canViewProject, err := o.CanViewProject(ctx, userId, projectUid, uIdOfPermission)
	if err != nil {
		return false, err
	}
	if canViewOperationRecord || canViewProject || isUserProjectAdmin {
		return true, nil
	}
	return false, nil
}

func (o *OpPermissionVerifyUsecase) HasManagePermission(ctx context.Context, userId, projectUid string, uIdOfPermission string) (bool, error) {
	canViewOperationRecord, err := o.HasOpPermissionInProject(ctx, userId, projectUid, uIdOfPermission)
	if err != nil {
		return false, err
	}
	isUserProjectAdmin, err := o.IsUserProjectAdmin(ctx, userId, projectUid)
	if err != nil {
		return false, err
	}
	canOpProject, err := o.CanOpProject(ctx, userId, projectUid)
	if err != nil {
		return false, err
	}
	if canViewOperationRecord || canOpProject || isUserProjectAdmin {
		return true, nil
	}
	return false, nil
}

// HasOpPermissionInProject 查看某用户在某项目下是否有某种权限
func (o *OpPermissionVerifyUsecase) HasOpPermissionInProject(ctx context.Context, userUid, projectUid string, permissionUid string) (bool, error) {
	has, err := o.repo.IsUserHasOpPermissionInProject(ctx, userUid, projectUid, permissionUid)
	if err != nil {
		return false, fmt.Errorf("failed to check user op permission in project: %v", err)
	}
	return has, nil
}

func (o *OpPermissionVerifyUsecase) CanViewGlobal(ctx context.Context, userUid string) (bool, error) {
	isUserDMSAdmin, err := o.IsUserDMSAdmin(ctx, userUid)
	if err != nil {
		return false, err
	}
	if isUserDMSAdmin {
		return true, nil
	}

	ops, err := o.repo.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return false, err
	}

	for _, op := range ops {
		if op.UID == pkgConst.UIDOfOpPermissionGlobalManagement || op.UID == pkgConst.UIDOfOpPermissionGlobalView {
			return true, nil
		}
	}

	return false, nil
}

type OpPermissionWithOpRange struct {
	OpPermissionUID string      // 操作权限
	OpRangeType     OpRangeType // OpRangeType描述操作权限的权限范围类型，目前只支持数据源
	RangeUIDs       []string    // Range描述操作权限的权限范围，如涉及哪些数据源
}

func (o *OpPermissionVerifyUsecase) GetUserGlobalOpPermission(ctx context.Context, userUid string) ([]OpPermissionWithOpRange, error) {
	opPermissionWithOpRanges, err := o.repo.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user global op permission : %v", err)
	}

	var opPermissionWithOpRangesResult []OpPermissionWithOpRange
	for _, permission := range opPermissionWithOpRanges {
		opPermissionWithOpRangesResult = append(opPermissionWithOpRangesResult, OpPermissionWithOpRange{
			OpPermissionUID: permission.UID,
			OpRangeType:     permission.RangeType,
		})
	}

	return opPermissionWithOpRangesResult, nil
}

func (o *OpPermissionVerifyUsecase) GetUserOpPermissionInProject(ctx context.Context, userUid, projectUid string) ([]OpPermissionWithOpRange, error) {

	opPermissionWithOpRanges, err := o.repo.GetUserOpPermissionInProject(ctx, userUid, projectUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user op permission in project: %v", err)
	}
	opProjectPermissionWithOpRanges, err := o.repo.GetUserProjectOpPermissionInProject(ctx, userUid, projectUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project op permission: %v", err)
	}
	opPermissionWithOpRanges = append(opPermissionWithOpRanges, opProjectPermissionWithOpRanges...)

	return opPermissionWithOpRanges, nil
}

func (o *OpPermissionVerifyUsecase) GetUserOpPermission(ctx context.Context, userUid string) ([]OpPermissionWithOpRange, error) {
	opPermissionWithOpRanges, err := o.repo.GetUserOpPermission(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user op permission in project: %v", err)
	}
	opProjectPermissionWithOpRanges, err := o.repo.GetUserProjectOpPermission(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project op permission: %v", err)
	}
	opPermissionWithOpRanges = append(opPermissionWithOpRanges, opProjectPermissionWithOpRanges...)
	return opPermissionWithOpRanges, nil
}

type ProjectOpPermissionWithOpRange struct {
	ProjectUid              string
	ProjectName             string
	OpPermissionWithOpRange OpPermissionWithOpRange
}

func (o *OpPermissionVerifyUsecase) GetUserProjectOpPermission(ctx context.Context, userUid string) ([]ProjectOpPermissionWithOpRange, error) {

	projectOpPermissionWithOpRange, err := o.repo.GetUserProjectWithOpPermissions(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project with op permission : %v", err)
	}

	return projectOpPermissionWithOpRange, nil
}

func (o *OpPermissionVerifyUsecase) GetUserManagerProject(ctx context.Context, projectWithOpPermissions []ProjectOpPermissionWithOpRange) (userBindProjects []dmsCommonV1.UserBindProject) {

	/* 结果如下，需要去重
	+--------+---------+-------------------+---------------+---------------------+
	| uid    | name    | op_permission_uid | op_range_type | range_uids          |
	+--------+---------+-------------------+---------------+---------------------+
	| 700300 | default | 700003            | db_service    | 1650760484527280128 |
	+--------+---------+-------------------+---------------+---------------------+
	| 700300 |	default| 700002	 		   | project	   |	700300			 |
	+--------+---------+-------------------+---------------+---------------------+
	*/
	mapIdUserBindProject := make(map[string]dmsCommonV1.UserBindProject, 0)
	for _, projectWithOpPermission := range projectWithOpPermissions {
		n, ok := mapIdUserBindProject[projectWithOpPermission.ProjectUid]
		if !ok {
			mapIdUserBindProject[projectWithOpPermission.ProjectUid] = dmsCommonV1.UserBindProject{ProjectID: projectWithOpPermission.ProjectUid, ProjectName: projectWithOpPermission.ProjectName, IsManager: projectWithOpPermission.OpPermissionWithOpRange.OpPermissionUID == pkgConst.UIDOfOpPermissionProjectAdmin}
		} else {
			// 有一个权限为项目管理员即可
			n.IsManager = mapIdUserBindProject[projectWithOpPermission.ProjectUid].IsManager || (projectWithOpPermission.OpPermissionWithOpRange.OpPermissionUID == pkgConst.UIDOfOpPermissionProjectAdmin)
			mapIdUserBindProject[projectWithOpPermission.ProjectUid] = n
		}
	}

	for _, userBindProject := range mapIdUserBindProject {
		userBindProjects = append(userBindProjects, userBindProject)
	}

	return userBindProjects
}

func (o *OpPermissionVerifyUsecase) CanCreateProject(ctx context.Context, userUid string) (bool, error) {
	// user admin has all op permission
	hasGlobalOpPermission, err := o.CanOpGlobal(ctx, userUid)
	if err != nil {
		return false, err
	}
	if hasGlobalOpPermission {
		return true, nil
	}

	opPermissions, err := o.repo.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return false, fmt.Errorf("failed to get user global op permission : %v", err)
	}
	for _, opPermission := range opPermissions {
		if opPermission.UID == pkgConst.UIDOfOpPermissionCreateProject {
			return true, nil
		}
	}

	return false, nil
}

type ListMembersOpPermissionOption struct {
	PageNumber   uint32
	LimitPerPage uint32
}

type ListMembersOpPermissionItem struct {
	UserUid       string
	UserName      string
	OpPermissions []OpPermissionWithOpRange
}

func (o *OpPermissionVerifyUsecase) ListUsersOpPermissionInProject(ctx context.Context, projectUid string, opt *ListMembersOpPermissionOption) ([]ListMembersOpPermissionItem, int64, error) {

	items, total, err := o.repo.ListUsersOpPermissionInProject(ctx, projectUid, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list members op permission in project: %v", err)
	}

	return items, total, nil
}

func (o *OpPermissionVerifyUsecase) ListUsersInProject(ctx context.Context, projectUid string) ([]ListMembersOpPermissionItem, error) {
	return o.repo.ListUsersInProject(ctx, projectUid)
}

func (o *OpPermissionVerifyUsecase) GetUserProject(ctx context.Context, userUid string) ([]*Project, error) {

	projects, err := o.repo.GetUserProject(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project with op permission : %v", err)
	}

	return projects, nil
}

func (o *OpPermissionVerifyUsecase) UserCanOpDB(userOpPermissions []OpPermissionWithOpRange, needOpPermissionTypes []string, dbServiceUid string) bool {
	for _, userOpPermission := range userOpPermissions {
		// 项目管理员可以查看所有数据源
		if userOpPermission.OpPermissionUID == pkgConst.UIDOfOpPermissionProjectAdmin {
			return true
		}

		// 动作权限(创建、审核、上线工单等)
		for _, needOpPermission := range needOpPermissionTypes {
			if needOpPermission == userOpPermission.OpPermissionUID && userOpPermission.OpRangeType == OpRangeType(dmsV1.OpRangeTypeDBService) {
				// 对象权限(指定数据源)
				for _, id := range userOpPermission.RangeUIDs {
					if id == dbServiceUid {
						return true
					}
				}
			}
		}
	}

	return false
}

func (o *OpPermissionVerifyUsecase) GetCanOpDBUsers(ctx context.Context, projectUID, dbServiceUid string, needOpPermissionTypes []string) ([]string, error) {
	members, _, err := o.ListUsersOpPermissionInProject(ctx, projectUID, &ListMembersOpPermissionOption{
		PageNumber:   1,
		LimitPerPage: 999,
	})
	if nil != err {
		return nil, err
	}

	userIds := make([]string, 0)
	for _, member := range members {
		if o.UserCanOpDB(member.OpPermissions, needOpPermissionTypes, dbServiceUid) {
			userIds = append(userIds, member.UserUid)
		}
	}

	return userIds, nil
}
