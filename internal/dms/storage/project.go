package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.ProjectRepo = (*ProjectRepo)(nil)

type ProjectRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewProjectRepo(log utilLog.Logger, s *Storage) *ProjectRepo {
	return &ProjectRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.project"))}
}

func (d *ProjectRepo) SaveProject(ctx context.Context, u *biz.Project) error {
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(convertBizProject(u)).Error; err != nil {
			return fmt.Errorf("failed to save project: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *ProjectRepo) BatchSaveProjects(ctx context.Context, projects []*biz.Project) error {
	models := make([]*model.Project, 0)
	for _, project := range projects {
		models = append(models, convertBizProject(project))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).CreateInBatches(models, 50).Error; err != nil {
			return fmt.Errorf("failed to batch save projects: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// TODO currentUserUid 未生效
func (d *ProjectRepo) ListProjects(ctx context.Context, opt *biz.ListProjectsOption, currentUserUid string) (projects []*biz.Project, total int64, err error) {
	var models []*model.Project

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list projects: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.Project{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count projects: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelProject(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model projects: %v", err))
		}
		// ds.CreateUserName = model.UserName
		projects = append(projects, ds)
	}
	return projects, total, nil
}

func (d *ProjectRepo) GetProject(ctx context.Context, projectUid string) (*biz.Project, error) {
	var project *model.Project
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&project, "uid = ?", projectUid).Error; err != nil {
			return fmt.Errorf("failed to get project: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelProject(project)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model project: %v", err))
	}
	return ret, nil
}

func (d *ProjectRepo) GetProjectByName(ctx context.Context, projectName string) (*biz.Project, error) {
	var project *model.Project
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&project, "name = ?", projectName).Error; err != nil {
			return fmt.Errorf("failed to get project by name: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelProject(project)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model project: %v", err))
	}
	return ret, nil
}

func (d *ProjectRepo) GetProjectByNames(ctx context.Context, projectNames []string) ([]*biz.Project, error) {
	var projects []*model.Project
	bizProjects := make([]*biz.Project, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Where("name IN (?)", projectNames).Find(&projects).Error; err != nil {
			return fmt.Errorf("failed to get project by name: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for _, project := range projects {
		bizProject, err := convertModelProject(project)
		if err != nil {
			continue
		}
		bizProjects = append(bizProjects, bizProject)
	}
	return bizProjects, nil
}

func (d *ProjectRepo) UpdateProject(ctx context.Context, u *biz.Project) error {
	_, err := d.GetProject(ctx, u.UID)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("project not exist"))
		}
		return err
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Project{}).Where("uid = ?", u.UID).Omit("created_at").Save(convertBizProject(u)).Error; err != nil {
			return fmt.Errorf("failed to update project: %v", err)
		}
		return nil
	})

}

func (d *ProjectRepo) DelProject(ctx context.Context, projectUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", projectUid).Delete(&model.Project{}).Error; err != nil {
			return fmt.Errorf("failed to delete project: %v", err)
		}
		return nil
	})
}

func (d *ProjectRepo) UpdateDBServiceBusiness(ctx context.Context, projectUid string, originBusiness string, descBusiness string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DBService{}).Where("project_uid = ? and business = ?", projectUid, originBusiness).
			Update("business", descBusiness).Error; err != nil {
			return fmt.Errorf("failed to update dbService business: %v", err)
		}
		return nil
	})
}

func (d *ProjectRepo) IsProjectActive(ctx context.Context, projectUid string) error {
	project, err := d.GetProject(ctx, projectUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("project not exist"))
		}
		return err
	}

	if project.Status != biz.ProjectStatusActive {
		return fmt.Errorf("project status is : %v", project.Status)
	}
	return nil
}
