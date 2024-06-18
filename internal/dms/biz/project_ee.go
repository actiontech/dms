//go:build enterprise

package biz

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/gocarina/gocsv"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *ProjectUsecase) CreateProject(ctx context.Context, project *Project, createUserUID string) (err error) {
	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	err = d.repo.SaveProject(tx, project)
	if err != nil {
		return fmt.Errorf("save projects failed: %v", err)
	}

	// 默认将admin用户加入空间成员，并且为管理员
	_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, pkgConst.UIDOfUserAdmin, project.UID)
	if err != nil {
		return fmt.Errorf("add admin to projects failed: %v", err)
	}
	// 非admin用户创建时,默认将空间创建人加入空间成员，并且为管理员
	if createUserUID != pkgConst.UIDOfUserAdmin {
		_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, createUserUID, project.UID)
		if err != nil {
			return fmt.Errorf("add create user to projects failed: %v", err)
		}
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	// plugin handle after create project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, project.UID, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeCreate, dmsV1.OperationTimingAfter)
	if err != nil {
		return fmt.Errorf("plugin handle after create project failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) GetProjectByName(ctx context.Context, projectName string) (*Project, error) {
	return d.repo.GetProjectByName(ctx, projectName)
}

func (d *ProjectUsecase) UpdateProjectDesc(ctx context.Context, currentUserUid, projectUid string, desc *string) (err error) {
	if err := d.checkUserCanUpdateProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("user can't update project: %v", err)
	}

	project, err := d.repo.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project err: %v", err)
	}

	if desc != nil {
		project.Desc = *desc
	}

	err = d.repo.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("update projects desc failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) ArchivedProject(ctx context.Context, currentUserUid, projectUid string, archived bool) (err error) {
	if projectUid == pkgConst.UIDOfProjectDefault {
		return fmt.Errorf("default project is not allow to archive")
	}
	if err := d.checkUserCanUpdateProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("user can't update project: %v", err)
	}

	project, err := d.repo.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project err: %v", err)
	}

	// 调整空间状态
	var status ProjectStatus
	if archived {
		status = ProjectStatusArchived
	} else {
		status = ProjectStatusActive
	}
	if status == project.Status {
		return fmt.Errorf("can't operate project current status is %v", status)
	}
	project.Status = status

	// plugin check before delete project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
	if err != nil {
		return fmt.Errorf("check before delete project failed: %v", err)
	}

	err = d.repo.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("update projects status failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) DeleteProject(ctx context.Context, currentUserUid, projectUid string) (err error) {
	// check
	{
		if projectUid == pkgConst.UIDOfProjectDefault {
			return fmt.Errorf("default project is not allow to delete")
		}
		// project admin can delete project
		isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid)
		if err != nil {
			return fmt.Errorf("check user project admin error: %v", err)
		}
		if !isAdmin {
			return fmt.Errorf("user can't update project")
		}

		// plugin check before delete project
		err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
		if err != nil {
			return fmt.Errorf("check before delete project failed: %v", err)
		}

	}
	err = d.repo.DelProject(ctx, projectUid)
	if err != nil {
		return err
	}
	// plugin clean unused data after delete project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingAfter)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectUsecase) checkUserCanUpdateProject(ctx context.Context, currentUserUid, projectUid string) error {
	// project admin can update project
	isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid)
	if err != nil {
		return fmt.Errorf("check user project admin error: %v", err)
	}
	if !isAdmin {
		return fmt.Errorf("user can't update project")
	}
	return nil
}

func (d *ProjectUsecase) isProjectActive(ctx context.Context, projectUid string) error {
	project, err := d.GetProject(ctx, projectUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("project not exist"))
		}
		return err
	}

	if project.Status != ProjectStatusActive {
		return fmt.Errorf("project status is : %v", project.Status)
	}
	return nil
}

func (d *ProjectUsecase) ImportProjects(ctx context.Context, uid string, projects []*Project) error {
	if len(projects) == 0 {
		return fmt.Errorf("projects is empty")
	}

	err := d.checkImportOp(ctx, uid, projects)
	if err != nil {
		return fmt.Errorf("check import op failed: %w", err)
	}

	err = d.batchSaveProjects(ctx, uid, projects)
	if err != nil {
		return fmt.Errorf("batch save projects failed: %w", err)
	}

	return nil
}

func (d *ProjectUsecase) batchSaveProjects(ctx context.Context, uid string, projects []*Project) (err error) {
	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err = d.repo.BatchSaveProjects(tx, projects); err != nil {
		return fmt.Errorf("failed to batch save projects: %v", err)
	}

	for _, project := range projects {
		// 默认将admin用户加入空间成员，并且为管理员
		_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, pkgConst.UIDOfUserAdmin, project.UID)
		if err != nil {
			return fmt.Errorf("add admin to projects failed: %v", err)
		}

		// 非admin用户创建时,默认将空间创建人加入空间成员，并且为管理员
		if uid != pkgConst.UIDOfUserAdmin {
			_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, uid, project.UID)
			if err != nil {
				return fmt.Errorf("add create user to projects failed: %v", err)
			}
		}

		err = d.pluginUsecase.OperateDataResourceHandle(ctx, project.UID, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeCreate, dmsV1.OperationTimingAfter)
		if err != nil {
			return fmt.Errorf("plugin handle after create project failed: %v", err)
		}
	}

	err = tx.Commit(d.log)
	if err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) checkImportOp(ctx context.Context, uid string, projects []*Project) error {
	if len(projects) == 0 {
		return fmt.Errorf("projects is empty")
	}

	// check current user has enough permission
	if canCreateProject, err := d.opPermissionVerifyUsecase.CanCreateProject(ctx, uid); err != nil {
		return fmt.Errorf("failed to check user can create project: %v", err)
	} else if !canCreateProject {
		return fmt.Errorf("current user %s can't create project", uid)
	}

	m := make(map[string]struct{})
	projectNames := make([]string, 0)
	for _, p := range projects {
		if _, ok := m[p.Name]; !ok {
			// 因为业务名称是以分号分割的，所以项目名称不能包含分号
			if strings.Contains(p.Name, ";") {
				return fmt.Errorf("project name %s contains split character ';',please check file content", p.Name)
			}

			m[p.Name] = struct{}{}
			projectNames = append(projectNames, p.Name)
		} else {
			return fmt.Errorf("project name %s is duplicate,please check file content", p.Name)
		}
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	filterBy = append(filterBy, pkgConst.FilterCondition{
		Field:    string(ProjectFieldName),
		Operator: pkgConst.FilterOperatorIn,
		Value:    projectNames,
	})

	// 找出已经存在的项目
	_, total, err := d.repo.ListProjects(ctx, &ListProjectsOption{
		PageNumber:   1,
		LimitPerPage: 999999,
		FilterBy:     filterBy,
	}, uid)
	if err != nil {
		return fmt.Errorf("failed to list projects: %v", err)
	}

	if total > 0 {
		return fmt.Errorf("project name is same with existing project,please check file content")
	}

	return nil
}

//go:embed template/import_projects_template.csv
var importTemplate []byte

func (d *ProjectUsecase) GetImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	if canCreateProject, err := d.opPermissionVerifyUsecase.CanCreateProject(ctx, uid); err != nil {
		return nil, err
	} else if !canCreateProject {
		return nil, fmt.Errorf("current user can't create project")
	}

	return importTemplate, nil
}

func (d *ProjectUsecase) GetProjectTips(ctx context.Context, uid, projectUid string) ([]*Project, error) {
	filterBy := make([]pkgConst.FilterCondition, 0)
	if projectUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(ProjectFieldUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    projectUid,
		})
	}

	listOption := &ListProjectsOption{
		PageNumber:   1,
		LimitPerPage: 99999,
		FilterBy:     filterBy,
	}

	projects, _, err := d.ListProject(ctx, listOption, uid)
	if nil != err {
		return nil, err
	}

	return projects, nil
}

func (d *ProjectUsecase) PreviewImportProjects(ctx context.Context, uid, file string) ([]*PreviewProject, error) {
	if len(file) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// check current user has enough permission
	if canCreateProject, err := d.opPermissionVerifyUsecase.CanCreateProject(ctx, uid); err != nil {
		return nil, fmt.Errorf("failed to check user can create project: %v", err)
	} else if !canCreateProject {
		return nil, fmt.Errorf("current user %s can't create project", uid)
	}

	r := csv.NewReader(bytes.NewReader([]byte(file)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %v", err)
	}

	projects := make([]*PreviewProject, 0, len(records)-1)
	for i, record := range records {
		// 跳过表头
		if i == 0 {
			continue
		}

		if len(record) != 3 {
			return nil, fmt.Errorf("invalid record length: %d", len(record))
		}

		projects = append(projects, &PreviewProject{
			Name:     record[0],
			Desc:     record[1],
			Business: strings.Split(record[2], ";"),
		})
	}

	return projects, nil
}

func (d *ProjectUsecase) ExportProjects(ctx context.Context, uid string, option *ListProjectsOption) ([]byte, error) {
	if canCreateProject, err := d.opPermissionVerifyUsecase.CanCreateProject(ctx, uid); err != nil {
		return nil, fmt.Errorf("failed to check user can create project: %v", err)
	} else if !canCreateProject {
		return nil, fmt.Errorf("current user %s can't create project", uid)
	}

	projects, _, err := d.repo.ListProjects(ctx, option, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}

	buff := new(bytes.Buffer)
	buff.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	csvWriter := csv.NewWriter(buff)

	if err := csvWriter.Write([]string{
		"项目名称",
		"项目描述",
		"项目状态",
		"可用业务",
		"创建时间",
	}); err != nil {
		return nil, fmt.Errorf("failed to write csv header: %v", err)
	}

	for _, project := range projects {
		var status string
		if project.Status == ProjectStatusArchived {
			status = "不可用"
		} else {
			status = "可用"
		}

		var business string
		for _, b := range project.Business {
			business += b.Name + ";"
		}

		if err := csvWriter.Write([]string{
			project.Name,
			project.Desc,
			status,
			business,
			project.CreateTime.Format("2006-01-02 15:04:05"),
		}); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %v", err)
		}
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush csv writer: %v", err)
	}

	return buff.Bytes(), nil
}

func (d *ProjectUsecase) UpdateProject(ctx context.Context, currentUserUid, projectUid string, desc *string, isFixBusiness *bool, business []v1.Business) (err error) {
	if err := d.checkUserCanUpdateProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("user can't update project: %v", err)
	}

	project, err := d.repo.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project err: %v", err)
	}

	if desc != nil {
		project.Desc = *desc
	}

	if isFixBusiness != nil {
		project.IsFixedBusiness = *isFixBusiness
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if business != nil {
		businessList := make([]Business, 0)
		for _, b := range business {
			var oldBusiness string
			newBusiness := b.Name
			uid := b.ID
			if b.ID == "" {
				uid, err = pkgRand.GenStrUid()
				if err != nil {
					return fmt.Errorf("gen business uid failed: %v", err)
				}
			} else {
				pj := new(Project)
				pj, err = d.repo.GetProject(tx, projectUid)
				if err != nil {
					return fmt.Errorf("get business by uid failed: %v", err)
				}
				for _, bs := range pj.Business {
					if bs.Uid == b.ID {
						oldBusiness = bs.Name
					}
				}

				err = d.UpdateDBServiceBusiness(tx, currentUserUid, project.UID, oldBusiness, newBusiness)
				if err != nil {
					return fmt.Errorf("update db service business failed: %v", err)
				}
			}

			businessList = append(businessList, Business{
				Uid:  uid,
				Name: newBusiness,
			})
		}

		project.Business = businessList
	}

	err = d.repo.UpdateProject(tx, project)
	if err != nil {
		return fmt.Errorf("update projects desc failed: %v", err)
	}

	if err = tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}

var importDBServicesTemplateData []byte

func init() {
	var importDBServicesTemplateRows = []*ImportDbServicesCsvRow{
		{
			DbName:           "mysql_1",
			ProjName:         "default",
			Business:         "test",
			Desc:             "mysql_1",
			DbType:           "MySQL",
			Host:             "127.0.0.1",
			Port:             "3306",
			User:             "root",
			Password:         "123456",
			OracleService:    "",
			DB2DbName:        "",
			OpsTime:          "22:30-23:59;00:00-06:00",
			RuleTemplateName: "default_MySQL",
			AuditLevel:       "notice",
		}, {
			DbName:           "oracle_1",
			ProjName:         "default",
			Business:         "test",
			Desc:             "oracle_1",
			DbType:           "Oracle",
			Host:             "127.0.0.1",
			Port:             "1521",
			User:             "system",
			Password:         "123456",
			OracleService:    "xe",
			DB2DbName:        "",
			OpsTime:          "",
			RuleTemplateName: "default_Oracle",
			AuditLevel:       "warn",
		}, {
			DbName:           "db2_1",
			ProjName:         "default",
			Business:         "test",
			Desc:             "db2_1",
			DbType:           "DB2",
			Host:             "127.0.0.1",
			Port:             "50000",
			User:             "db2inst1",
			Password:         "123456",
			OracleService:    "",
			DB2DbName:        "testdb",
			OpsTime:          "9:30-11:30;13:10-18:10",
			RuleTemplateName: "default_DB2",
			AuditLevel:       "error",
		},
	}

	data, err := gocsv.MarshalBytes(importDBServicesTemplateRows)
	if err != nil {
		panic(err)
	}
	importDBServicesTemplateData = data
}

func (d *ProjectUsecase) GetImportDBServicesTemplate(ctx context.Context, uid string) ([]byte, error) {
	return importDBServicesTemplateData, nil
}
