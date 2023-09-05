package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.OpPermissionVerifyRepo = (*OpPermissionVerifyRepo)(nil)

type OpPermissionVerifyRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOpPermissionVerifyRepo(log utilLog.Logger, s *Storage) *OpPermissionVerifyRepo {
	return &OpPermissionVerifyRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.op_permission_verify"))}
}

func (o *OpPermissionVerifyRepo) IsUserHasOpPermissionInNamespace(ctx context.Context, userUid, namespaceUid, opPermissionUid string) (has bool, err error) {
	var count int64
	var memberGroupCount int64
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    count(*) 
		FROM members AS m 
		JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid=? AND m.namespace_uid=? 
		JOIN role_op_permissions AS p ON r.role_uid = p.role_uid AND p.op_permission_uid = ?`, userUid, namespaceUid, opPermissionUid).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in namespace: %v", err)
		}

		if err := tx.WithContext(ctx).Raw(`
		SELECT 
    		count(*) 
		FROM member_groups AS mg
		JOIN member_group_users AS mgu ON mg.uid = mgu.member_group_uid and mgu.user_uid = ? AND mg.namespace_uid = ? 
		JOIN member_group_role_op_ranges AS mgrop ON mg.uid = mgrop.member_group_uid 
		JOIN role_op_permissions AS rop ON mgrop.role_uid = rop.role_uid AND rop.op_permission_uid = ?`, userUid, namespaceUid, opPermissionUid).Count(&memberGroupCount).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in namespace: %v", err)
		}

		return nil
	}); err != nil {
		return false, err
	}
	return count > 0 || memberGroupCount > 0, nil
}
func (o *OpPermissionVerifyRepo) GetUserNamespaceWithOpPermissions(ctx context.Context, userUid string) (namespaceWithPermission []biz.NamespaceOpPermissionWithOpRange, err error) {
	var ret []struct {
		Uid             string
		Name            string
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    n.uid, n.name, p.op_permission_uid, r.op_range_type, r.range_uids 
		FROM namespaces AS n
		JOIN members AS m ON n.uid = m.namespace_uid
		JOIN users AS u ON m.user_uid = u.uid AND u.uid = ?
		JOIN member_role_op_ranges AS r ON m.uid=r.member_uid
		JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
		WHERE n.status = 'active'
		UNION
		SELECT 
			distinct n.uid, n.name, rop.op_permission_uid, mgrop.op_range_type, mgrop.range_uids 
		FROM namespaces AS n
		JOIN member_groups AS mg ON n.uid = mg.namespace_uid
		JOIN member_group_users AS mgu ON mg.uid = mgu.member_group_uid
		JOIN users AS u ON mgu.user_uid = u.uid AND u.uid = ?
		JOIN member_group_role_op_ranges AS mgrop ON mg.uid=mgrop.member_group_uid
		JOIN role_op_permissions AS rop ON mgrop.role_uid = rop.role_uid
		WHERE n.status = 'active'
	`, userUid, userUid).Find(&ret).Error; err != nil {
			return fmt.Errorf("failed to find user op permission with namespace: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	namespaceWithPermission = make([]biz.NamespaceOpPermissionWithOpRange, 0, len(ret))
	for _, r := range ret {
		typ, err := biz.ParseOpRangeType(r.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse op range type: %v", err)
		}

		namespaceWithPermission = append(namespaceWithPermission, biz.NamespaceOpPermissionWithOpRange{
			NamespaceUid:  r.Uid,
			NamespaceName: r.Name,
			OpPermissionWithOpRange: biz.OpPermissionWithOpRange{
				OpPermissionUID: r.OpPermissionUid,
				OpRangeType:     typ,
				RangeUIDs:       convertModelRangeUIDs(r.RangeUids),
			},
		})
	}
	return namespaceWithPermission, nil
}

func (o *OpPermissionVerifyRepo) GetUserOpPermissionInNamespace(ctx context.Context, userUid, namespaceUid string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    p.op_permission_uid, r.op_range_type, r.range_uids 
		FROM members AS m 
		JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid=? AND m.namespace_uid=? 
		JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
		UNION 
		SELECT
			DISTINCT rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
		FROM member_groups mg
		JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
		JOIN member_group_role_op_ranges mgror ON mgu.member_group_uid = mgror.member_group_uid
		JOIN role_op_permissions rop ON mgror.role_uid = rop.role_uid
		WHERE mg.namespace_uid = ? and mgu.user_uid = ?`, userUid, namespaceUid, namespaceUid, userUid).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to get user op permission in namespace: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	opPermissionWithOpRanges = make([]biz.OpPermissionWithOpRange, 0, len(results))
	for _, r := range results {
		typ, err := biz.ParseOpRangeType(r.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse op range type: %v", err)
		}

		opPermissionWithOpRanges = append(opPermissionWithOpRanges, biz.OpPermissionWithOpRange{
			OpPermissionUID: r.OpPermissionUid,
			OpRangeType:     typ,
			RangeUIDs:       convertModelRangeUIDs(r.RangeUids),
		})
	}

	return opPermissionWithOpRanges, nil
}

func (o *OpPermissionVerifyRepo) GetUserOpPermission(ctx context.Context, userUid string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	if err = transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err = tx.WithContext(ctx).Raw(`
		SELECT 
			p.op_permission_uid, r.op_range_type, r.range_uids 
		FROM members AS m 
		JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid = ?
		JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
		UNION 
		select 
			distinct rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
		from member_group_users mgu
		join member_group_role_op_ranges mgror on mgu.member_group_uid = mgror.member_group_uid
		join role_op_permissions rop on mgror.role_uid = rop.role_uid
		where mgu.user_uid = ?`, userUid, userUid).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to get user op permission: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	opPermissionWithOpRanges = make([]biz.OpPermissionWithOpRange, 0, len(results))
	for _, r := range results {
		typ, err := biz.ParseOpRangeType(r.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse op range type: %v", err)
		}

		opPermissionWithOpRanges = append(opPermissionWithOpRanges, biz.OpPermissionWithOpRange{
			OpPermissionUID: r.OpPermissionUid,
			OpRangeType:     typ,
			RangeUIDs:       convertModelRangeUIDs(r.RangeUids),
		})
	}

	return opPermissionWithOpRanges, nil
}

func (o *OpPermissionVerifyRepo) GetUserGlobalOpPermission(ctx context.Context, userUid string) (opPermissions []*biz.OpPermission, err error) {
	ops := []*model.OpPermission{}
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		user := model.User{Model: model.Model{UID: userUid}}
		if err := tx.WithContext(ctx).Model(&user).Association("OpPermissions").Find(&ops); err != nil {
			return fmt.Errorf("failed to get user global op permission: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	opPermissions = make([]*biz.OpPermission, 0, len(ops))
	for _, o := range ops {
		opPermission, err := convertModelOpPermission(o)
		if err != nil {
			return nil, err
		}
		opPermissions = append(opPermissions, opPermission)
	}

	return opPermissions, nil
}

func (o *OpPermissionVerifyRepo) ListUsersOpPermissionInNamespace(ctx context.Context, namespaceUid string, opt *biz.ListMembersOpPermissionOption) (items []biz.ListMembersOpPermissionItem, total int64, err error) {
	type result struct {
		MemberUid       string
		MemberGroupUid  string
		UserUid         string
		UserName        string
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	var permissionResults []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		// opt中的分页属性作用于空间内的成员，即opt.LimitPerPage表示返回opt.LimitPerPage个成员的权限

		// find result
		{
			if err := tx.WithContext(ctx).Raw(`
			SELECT * FROM (
				SELECT 
					m.uid AS member_uid, IFNULL(NULL, "") member_group_uid, m.user_uid, u.name AS user_name 
				FROM
					members AS m JOIN users AS u ON m.user_uid = u.uid AND m.namespace_uid = ?
				UNION
				SELECT 
					IFNULL(NULL, "") AS member_uid, mg.uid AS member_group_uid, u.uid AS user_uid, u.name AS user_name
				FROM 
					member_groups AS mg
					JOIN member_group_users mgu on mg.uid = mgu.member_group_uid AND mg.namespace_uid = ?
					JOIN users AS u ON mgu.user_uid = u.uid
			) TEMP ORDER BY user_uid LIMIT ? OFFSET ?`,
				namespaceUid, namespaceUid, opt.LimitPerPage, opt.LimitPerPage*(uint32(fixPageIndices(opt.PageNumber)))).Scan(&results).Error; err != nil {
				return fmt.Errorf("failed to list user op permission in namespace: %v", err)
			}
		}

		// find total
		{
			if err := tx.WithContext(ctx).Raw(`
			SELECT COUNT(*) FROM (
				SELECT 
					m.uid AS member_uid, IFNULL(NULL, "") member_group_uid, m.user_uid, u.name AS user_name
				FROM members AS m 
				JOIN users AS u ON m.user_uid = u.uid AND m.namespace_uid=?
				UNION
				SELECT 
					IFNULL(NULL, "") AS member_uid, mg.uid AS member_group_uid, u.uid AS user_uid, u.name AS user_name 
				FROM member_groups AS mg
				JOIN member_group_users mgu on mg.uid = mgu.member_group_uid AND mg.namespace_uid = ?
				JOIN users AS u ON mgu.user_uid = u.uid
			) TEMP`,
				namespaceUid, namespaceUid).Scan(&total).Error; err != nil {
				return fmt.Errorf("failed to list total user op permission in namespace: %v", err)
			}
		}

		if len(results) > 0 {
			userIds := make([]string, 0)
			for _, item := range results {
				userIds = append(userIds, item.UserUid)
			}

			{
				if err = tx.WithContext(ctx).Raw(`
				SELECT 
					m.user_uid, p.op_permission_uid, r.op_range_type, r.range_uids 
				FROM members AS m 
				JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid in (?) AND m.namespace_uid=? 
				JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
				UNION 
				SELECT
					DISTINCT mgu.user_uid, rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
				FROM member_groups mg
				JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
				JOIN member_group_role_op_ranges mgror ON mgu.member_group_uid = mgror.member_group_uid
				JOIN role_op_permissions rop ON mgror.role_uid = rop.role_uid
				WHERE mg.namespace_uid = ? and mgu.user_uid in (?)`, userIds, namespaceUid, namespaceUid, userIds).Scan(&permissionResults).Error; err != nil {
					return fmt.Errorf("failed to get user op permission in namespace: %v", err)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, 0, err
	}

	resultGroupByUser := make(map[string][]result)
	for _, r := range permissionResults {
		resultGroupByUser[r.UserUid] = append(resultGroupByUser[r.UserUid], r)
	}

	for _, rs := range results {
		opPermissionWithOpRanges := make([]biz.OpPermissionWithOpRange, 0)

		if permissionItems, ok := resultGroupByUser[rs.UserUid]; ok {
			for _, r := range permissionItems {
				// 这里表示没有任何角色权限的成员
				if r.OpRangeType == "" {
					continue
				}
				typ, err := biz.ParseOpRangeType(r.OpRangeType)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to parse op range type: %v", err)
				}

				opPermissionWithOpRanges = append(opPermissionWithOpRanges, biz.OpPermissionWithOpRange{
					OpPermissionUID: r.OpPermissionUid,
					OpRangeType:     typ,
					RangeUIDs:       convertModelRangeUIDs(r.RangeUids),
				})
			}
		}

		items = append(items, biz.ListMembersOpPermissionItem{
			MemberUid:      rs.MemberUid,
			MemberGroupUid: rs.MemberGroupUid,
			UserUid:        rs.UserUid,
			UserName:       rs.UserName,
			OpPermissions:  opPermissionWithOpRanges,
		})
	}

	return items, total, nil
}
