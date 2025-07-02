package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
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

func (o *OpPermissionVerifyRepo) IsUserHasOpPermissionInProject(ctx context.Context, userUid, projectUid, opPermissionUid string) (has bool, err error) {
	var count int64
	var memberGroupCount int64
	var projectPermissionMemberCount int64
	var projectPermissionMemberGroupCount int64
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    count(*) 
		FROM members AS m 
		JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid=? AND m.project_uid=? 
		JOIN role_op_permissions AS p ON r.role_uid = p.role_uid AND p.op_permission_uid = ?`, userUid, projectUid, opPermissionUid).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in project: %v", err)
		}

		if err := tx.WithContext(ctx).Raw(`
		SELECT 
    		count(*) 
		FROM member_groups AS mg
		JOIN member_group_users AS mgu ON mg.uid = mgu.member_group_uid and mgu.user_uid = ? AND mg.project_uid = ? 
		JOIN member_group_role_op_ranges AS mgrop ON mg.uid = mgrop.member_group_uid 
		JOIN role_op_permissions AS rop ON mgrop.role_uid = rop.role_uid AND rop.op_permission_uid = ?`, userUid, projectUid, opPermissionUid).Count(&memberGroupCount).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in project: %v", err)
		}

		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    count(*) 
		FROM members AS m 
		JOIN member_op_permissions AS p ON m.uid = p.member_uid AND p.op_permission_uid = ?
		WHERE m.user_uid=? AND m.project_uid=?`,opPermissionUid,userUid,projectUid).Count(&projectPermissionMemberCount).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in project: %v", err)
		}

		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    count(*) 
		FROM member_groups AS mg 
		JOIN member_group_users AS mgu ON mg.uid = mgu.member_group_uid and mgu.user_uid = ? AND mg.project_uid = ? 
		JOIN member_group_op_permissions AS p ON mg.uid = p.member_group_uid AND p.op_permission_uid = ?`,userUid,projectUid,opPermissionUid).Count(&projectPermissionMemberGroupCount).Error; err != nil {
			return fmt.Errorf("failed to check user has op permission in project: %v", err)
		}

		return nil
	}); err != nil {
		return false, err
	}
	return count > 0 || memberGroupCount > 0 || projectPermissionMemberCount > 0 || projectPermissionMemberGroupCount > 0, nil
}
func (o *OpPermissionVerifyRepo) GetUserProjectWithOpPermissions(ctx context.Context, userUid string) (projectWithPermission []biz.ProjectOpPermissionWithOpRange, err error) {
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
		FROM projects AS n
		JOIN members AS m ON n.uid = m.project_uid
		JOIN users AS u ON m.user_uid = u.uid AND u.uid = ?
		LEFT JOIN member_role_op_ranges AS r ON m.uid=r.member_uid
		LEFT JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
		WHERE n.status = 'active'
		UNION
		SELECT 
			distinct n.uid, n.name, rop.op_permission_uid, mgrop.op_range_type, mgrop.range_uids 
		FROM projects AS n
		JOIN member_groups AS mg ON n.uid = mg.project_uid
		JOIN member_group_users AS mgu ON mg.uid = mgu.member_group_uid
		JOIN users AS u ON mgu.user_uid = u.uid AND u.uid = ?
		LEFT JOIN member_group_role_op_ranges AS mgrop ON mg.uid=mgrop.member_group_uid
		LEFT JOIN role_op_permissions AS rop ON mgrop.role_uid = rop.role_uid
		WHERE n.status = 'active'
	`, userUid, userUid).Find(&ret).Error; err != nil {
			return fmt.Errorf("failed to find user op permission with project: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	projectWithPermission = make([]biz.ProjectOpPermissionWithOpRange, 0, len(ret))
	for _, r := range ret {
		typ, err := biz.ParseOpRangeType(r.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse op range type: %v", err)
		}

		projectWithPermission = append(projectWithPermission, biz.ProjectOpPermissionWithOpRange{
			ProjectUid:  r.Uid,
			ProjectName: r.Name,
			OpPermissionWithOpRange: biz.OpPermissionWithOpRange{
				OpPermissionUID: r.OpPermissionUid,
				OpRangeType:     typ,
				RangeUIDs:       convertModelRangeUIDs(r.RangeUids),
			},
		})
	}
	return projectWithPermission, nil
}

func (o *OpPermissionVerifyRepo) GetUserOpPermissionInProject(ctx context.Context, userUid, projectUid string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    p.op_permission_uid, mror.op_range_type, mror.range_uids 
		FROM members AS m 
		JOIN member_role_op_ranges AS mror ON m.uid=mror.member_uid AND m.user_uid=? AND m.project_uid=? 
		JOIN role_op_permissions AS p ON mror.role_uid = p.role_uid
		JOIN roles AS r ON r.uid = p.role_uid AND r.stat = 0
		UNION 
		SELECT
			DISTINCT rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
		FROM member_groups mg
		JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
		JOIN member_group_role_op_ranges mgror ON mgu.member_group_uid = mgror.member_group_uid
		JOIN role_op_permissions rop ON mgror.role_uid = rop.role_uid
		JOIN roles AS r ON r.uid = rop.role_uid AND r.stat = 0
		WHERE mg.project_uid = ? and mgu.user_uid = ?`, userUid, projectUid, projectUid, userUid).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to get user op permission in project: %v", err)
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

func (o *OpPermissionVerifyRepo) GetOneOpPermissionInProject(ctx context.Context, userUid, projectUid, permissionId string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    p.op_permission_uid, mror.op_range_type, mror.range_uids 
		FROM members AS m 
		JOIN member_role_op_ranges AS mror ON m.uid=mror.member_uid AND m.user_uid=? AND m.project_uid=? 
		JOIN role_op_permissions AS p ON mror.role_uid = p.role_uid AND p.op_permission_uid=?
		JOIN roles AS r ON r.uid = p.role_uid AND r.stat = 0
		UNION 
		SELECT
			DISTINCT rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
		FROM member_groups mg
		JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
		JOIN member_group_role_op_ranges mgror ON mgu.member_group_uid = mgror.member_group_uid
		JOIN role_op_permissions rop ON mgror.role_uid = rop.role_uid AND rop.op_permission_uid=?
		JOIN roles AS r ON r.uid = rop.role_uid AND r.stat = 0
		WHERE mg.project_uid = ? and mgu.user_uid = ?`, userUid, projectUid, permissionId, permissionId, projectUid, userUid).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to get user op permission in project: %v", err)
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


func (o *OpPermissionVerifyRepo) GetUserProjectOpPermissionInProject(ctx context.Context, userUid, projectUid string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeDataSourceUids       string
	}
	var results []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    mop.op_permission_uid, 'project' as op_range_type, m.project_uid as range_uids
		FROM members AS m 
		JOIN member_op_permissions AS mop ON m.uid=mop.member_uid AND m.user_uid=? AND m.project_uid=? 
		UNION 
		SELECT
			DISTINCT mgop.op_permission_uid, 'project' as op_range_type, mg.project_uid as range_uids
		FROM member_groups mg
		JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
		JOIN member_group_op_permissions AS mgop ON mg.uid = mgop.member_group_uid
		WHERE mg.project_uid = ? and mgu.user_uid = ?`, userUid, projectUid, projectUid, userUid).Scan(&results).Error; err != nil {
			return fmt.Errorf("failed to get user op permission in project: %v", err)
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
			RangeUIDs:       convertModelRangeUIDs(r.RangeDataSourceUids),
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

func (o *OpPermissionVerifyRepo) GetUserProjectOpPermission(ctx context.Context, userUid string) (opPermissionWithOpRanges []biz.OpPermissionWithOpRange, err error) {
	type result struct {
		OpPermissionUid string
		OpRangeType     string
		RangeDataSourceUids       string
	}
	var results []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT 
		    mop.op_permission_uid, 'project' as op_range_type, m.project_uid as range_uids 
		FROM members AS m 
		JOIN member_op_permissions AS mop ON m.uid=mop.member_uid AND m.user_uid=?
		UNION 
		SELECT
			DISTINCT mgop.op_permission_uid, 'project' as op_range_type, mg.project_uid as range_uids
		FROM member_groups mg
		JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
		JOIN member_group_op_permissions AS mgop ON mg.uid = mgop.member_group_uid
		WHERE mgu.user_uid = ?`, userUid, userUid).Scan(&results).Error; err != nil {
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
			RangeUIDs:       convertModelRangeUIDs(r.RangeDataSourceUids),
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

func (o *OpPermissionVerifyRepo) ListUsersOpPermissionInProject(ctx context.Context, projectUid string, opt *biz.ListMembersOpPermissionOption) (items []biz.ListMembersOpPermissionItem, total int64, err error) {
	type result struct {
		UserUid         string
		UserName        string
		OpPermissionUid string
		OpRangeType     string
		RangeUids       string
	}
	var results []result
	var permissionResults []result
	if err := transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		// opt中的分页属性作用于项目内的成员，即opt.LimitPerPage表示返回opt.LimitPerPage个成员的权限

		// find result
		{
			if err := tx.WithContext(ctx).Raw(`
			SELECT * FROM (
				SELECT 
					m.user_uid, u.name AS user_name 
				FROM
					members AS m JOIN users AS u ON m.user_uid = u.uid AND m.project_uid = ?
				UNION
				SELECT 
					DISTINCT u.uid AS user_uid, u.name AS user_name
				FROM 
					member_groups AS mg
					JOIN member_group_users mgu on mg.uid = mgu.member_group_uid AND mg.project_uid = ?
					JOIN users AS u ON mgu.user_uid = u.uid
			) TEMP ORDER BY user_uid LIMIT ? OFFSET ?`,
				projectUid, projectUid, opt.LimitPerPage, opt.LimitPerPage*(uint32(fixPageIndices(opt.PageNumber)))).Scan(&results).Error; err != nil {
				return fmt.Errorf("failed to list user op permission in project: %v", err)
			}
		}

		// find total
		{
			if err := tx.WithContext(ctx).Raw(`
			SELECT COUNT(*) FROM (
				SELECT 
					m.user_uid, u.name AS user_name
				FROM members AS m 
				JOIN users AS u ON m.user_uid = u.uid AND m.project_uid=?
				UNION
				SELECT 
					DISTINCT u.uid AS user_uid, u.name AS user_name 
				FROM member_groups AS mg
				JOIN member_group_users mgu on mg.uid = mgu.member_group_uid AND mg.project_uid = ?
				JOIN users AS u ON mgu.user_uid = u.uid
			) TEMP`,
				projectUid, projectUid).Scan(&total).Error; err != nil {
				return fmt.Errorf("failed to list total user op permission in project: %v", err)
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
				JOIN member_role_op_ranges AS r ON m.uid=r.member_uid AND m.user_uid in (?) AND m.project_uid=? 
				JOIN role_op_permissions AS p ON r.role_uid = p.role_uid
				UNION 
				SELECT
					DISTINCT mgu.user_uid, rop.op_permission_uid, mgror.op_range_type, mgror.range_uids 
				FROM member_groups mg
				JOIN member_group_users mgu ON mg.uid = mgu.member_group_uid
				JOIN member_group_role_op_ranges mgror ON mgu.member_group_uid = mgror.member_group_uid
				JOIN role_op_permissions rop ON mgror.role_uid = rop.role_uid
				WHERE mg.project_uid = ? and mgu.user_uid in (?)`, userIds, projectUid, projectUid, userIds).Scan(&permissionResults).Error; err != nil {
					return fmt.Errorf("failed to get user op permission in project: %v", err)
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
			UserUid:       rs.UserUid,
			UserName:      rs.UserName,
			OpPermissions: opPermissionWithOpRanges,
		})
	}

	return items, total, nil
}

func (o *OpPermissionVerifyRepo) ListUsersInProject(ctx context.Context, projectUid string) (items []biz.ListMembersOpPermissionItem, err error) {
	type result struct {
		UserUid  string
		UserName string
	}
	var results []result
	if err = transaction(o.log, ctx, o.db, func(tx *gorm.DB) error {
		// find result
		{
			if err = tx.WithContext(ctx).Raw(`
			SELECT * FROM (
				SELECT 
					m.user_uid, u.name AS user_name 
				FROM
					members AS m JOIN users AS u ON m.user_uid = u.uid AND m.project_uid = ?
				UNION
				SELECT 
					DISTINCT u.uid AS user_uid, u.name AS user_name
				FROM 
					member_groups AS mg
					JOIN member_group_users mgu on mg.uid = mgu.member_group_uid AND mg.project_uid = ?
					JOIN users AS u ON mgu.user_uid = u.uid
			) TEMP`,
				projectUid, projectUid).Scan(&results).Error; err != nil {
				return fmt.Errorf("failed to list users in project: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, rs := range results {
		items = append(items, biz.ListMembersOpPermissionItem{
			UserUid:  rs.UserUid,
			UserName: rs.UserName,
		})
	}

	return items, nil
}

func (d *OpPermissionVerifyRepo) GetUserProject(ctx context.Context, userUid string) (projects []*biz.Project, err error) {
	var models []*model.Project
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
			SELECT 
			n.*
		FROM
			projects n
			JOIN members m ON n.uid = m.project_uid
			JOIN users u ON m.user_uid = u.uid AND u.uid = ?
		UNION
		SELECT 
			DISTINCT n.*
		FROM 
			projects n
			JOIN member_groups mg on n.uid = mg.project_uid
			JOIN member_group_users mgu ON mgu.member_group_uid = mg.uid
			JOIN users u ON mgu.user_uid = u.uid AND u.uid = ?
			`, userUid, userUid).Scan(&models).Error; err != nil {
			return fmt.Errorf("failed to list user project: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// convert model to biz
	for _, model := range models {
		ds, err := convertModelProject(model)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model projects: %v", err))
		}
		projects = append(projects, ds)
	}
	return projects, nil
}
