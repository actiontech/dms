package biz

import (
	"context"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OpPermissionVerifyRepo interface {
	IsUserHasOpPermissionInNamespace(ctx context.Context, userUid, namespaceUid, opPermissionUid string) (has bool, err error)
	GetUserOpPermissionInNamespace(ctx context.Context, userUid, namespaceUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserOpPermission(ctx context.Context, userUid string) (opPermissionWithOpRanges []OpPermissionWithOpRange, err error)
	GetUserGlobalOpPermission(ctx context.Context, userUid string) (opPermissions []*OpPermission, err error)
	GetUserNamespaceWithOpPermissions(ctx context.Context, userUid string) (namespaceWithPermission []NamespaceOpPermissionWithOpRange, err error)
	ListUsersOpPermissionInNamespace(ctx context.Context, namespaceUid string, opt *ListMembersOpPermissionOption) (items []ListMembersOpPermissionItem, total int64, err error)
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

func (o *OpPermissionVerifyUsecase) IsUserNamespaceAdmin(ctx context.Context, userUid, namespaceUid string) (bool, error) {
	// 内置用户admin和sys拥有所有权限
	switch userUid {
	case pkgConst.UIDOfUserAdmin, pkgConst.UIDOfUserSys:
		return true, nil
	default:
	}
	has, err := o.repo.IsUserHasOpPermissionInNamespace(ctx, userUid, namespaceUid, pkgConst.UIDOfOpPermissionNamespaceAdmin)
	if err != nil {
		return false, fmt.Errorf("failed to check user is namespace admin: %v", err)
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

type OpPermissionWithOpRange struct {
	OpPermissionUID string      // 操作权限
	OpRangeType     OpRangeType // OpRangeType描述操作权限的权限范围类型，目前只支持数据源
	RangeUIDs       []string    // Range描述操作权限的权限范围，如涉及哪些数据源
}

func (o *OpPermissionVerifyUsecase) GetUserOpPermissionInNamespace(ctx context.Context, userUid, namespaceUid string) ([]OpPermissionWithOpRange, error) {

	opPermissionWithOpRanges, err := o.repo.GetUserOpPermissionInNamespace(ctx, userUid, namespaceUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user op permission in namespace: %v", err)
	}

	return opPermissionWithOpRanges, nil
}

func (o *OpPermissionVerifyUsecase) GetUserOpPermission(ctx context.Context, userUid string) ([]OpPermissionWithOpRange, error) {
	opPermissionWithOpRanges, err := o.repo.GetUserOpPermission(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user op permission in namespace: %v", err)
	}

	return opPermissionWithOpRanges, nil
}

type NamespaceOpPermissionWithOpRange struct {
	NamespaceUid            string
	NamespaceName           string
	OpPermissionWithOpRange OpPermissionWithOpRange
}

func (o *OpPermissionVerifyUsecase) GetUserNamespaceOpPermission(ctx context.Context, userUid string) ([]NamespaceOpPermissionWithOpRange, error) {

	namespaceOpPermissionWithOpRange, err := o.repo.GetUserNamespaceWithOpPermissions(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user namespace with op permission : %v", err)
	}

	return namespaceOpPermissionWithOpRange, nil
}

func (o *OpPermissionVerifyUsecase) GetUserManagerNamespace(ctx context.Context, namespaceWithOpPermissions []NamespaceOpPermissionWithOpRange) (userBindNamespaces []dmsCommonV1.UserBindNamespace) {

	/* 结果如下，需要去重
	+--------+---------+-------------------+---------------+---------------------+
	| uid    | name    | op_permission_uid | op_range_type | range_uids          |
	+--------+---------+-------------------+---------------+---------------------+
	| 700300 | default | 700003            | db_service    | 1650760484527280128 |
	+--------+---------+-------------------+---------------+---------------------+
	| 700300 |	default| 700002	 		   | namespace	   |	700300			 |
	+--------+---------+-------------------+---------------+---------------------+
	*/
	mapIdUserBindNamespace := make(map[string]dmsCommonV1.UserBindNamespace, 0)
	for _, namespaceWithOpPermission := range namespaceWithOpPermissions {
		n, ok := mapIdUserBindNamespace[namespaceWithOpPermission.NamespaceUid]
		if !ok {
			mapIdUserBindNamespace[namespaceWithOpPermission.NamespaceUid] = dmsCommonV1.UserBindNamespace{NamespaceID: namespaceWithOpPermission.NamespaceUid, NamespaceName: namespaceWithOpPermission.NamespaceName, IsManager: namespaceWithOpPermission.OpPermissionWithOpRange.OpPermissionUID == pkgConst.UIDOfOpPermissionNamespaceAdmin}
		} else {
			// 有一个权限为空间管理员即可
			n.IsManager = mapIdUserBindNamespace[namespaceWithOpPermission.NamespaceUid].IsManager || (namespaceWithOpPermission.OpPermissionWithOpRange.OpPermissionUID == pkgConst.UIDOfOpPermissionNamespaceAdmin)
			mapIdUserBindNamespace[namespaceWithOpPermission.NamespaceUid] = n
		}
	}

	for _, userBindNamespace := range mapIdUserBindNamespace {
		userBindNamespaces = append(userBindNamespaces, userBindNamespace)
	}

	return userBindNamespaces
}

func (o *OpPermissionVerifyUsecase) CanCreateNamespace(ctx context.Context, userUid string) (bool, error) {
	// user admin has all op permission
	isUserDMSAdmin, err := o.IsUserDMSAdmin(ctx, userUid)
	if err != nil {
		return false, err
	}
	if isUserDMSAdmin {
		return true, nil
	}

	opPermissions, err := o.repo.GetUserGlobalOpPermission(ctx, userUid)
	if err != nil {
		return false, fmt.Errorf("failed to get user global op permission : %v", err)
	}
	for _, opPermission := range opPermissions {
		if opPermission.UID == pkgConst.UIDOfOpPermissionCreateNamespace {
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

func (o *OpPermissionVerifyUsecase) ListUsersOpPermissionInNamespace(ctx context.Context, namespaceUid string, opt *ListMembersOpPermissionOption) ([]ListMembersOpPermissionItem, int64, error) {

	items, total, err := o.repo.ListUsersOpPermissionInNamespace(ctx, namespaceUid, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list members op permission in namespace: %v", err)
	}

	return items, total, nil
}
