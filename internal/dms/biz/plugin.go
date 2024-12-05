package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type DMSPluginRepo interface {
	SavePlugin(ctx context.Context, u *Plugin) error
	UpdatePlugin(ctx context.Context, u *Plugin) error
	ListPlugins(ctx context.Context) ([]*Plugin, error)
}

type PluginUsecase struct {
	logger            utilLog.Logger
	repo              DMSPluginRepo
	registeredPlugins []*Plugin
}

type Plugin struct {
	Name                         string
	OperateDataResourceHandleUrl string
}

func (p *Plugin) String() string {
	return fmt.Sprintf("name=%v,OperateDataHandleUrl=%v",
		p.Name, p.OperateDataResourceHandleUrl)
}

func NewDMSPluginUsecase(logger utilLog.Logger, repo DMSPluginRepo) (*PluginUsecase, error) {
	plugins, err := repo.ListPlugins(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("list plugins from repo error: %v", err)
	}

	return &PluginUsecase{
		logger:            logger,
		repo:              repo,
		registeredPlugins: plugins,
	}, nil
}

func (p *PluginUsecase) RegisterPlugin(ctx context.Context, plugin *Plugin, currentUserUid string) error {
	log := utilLog.NewHelper(p.logger, utilLog.WithMessageKey("biz.dmsplugin"))

	// check if the user is sys
	if currentUserUid != pkgConst.UIDOfUserSys {
		return fmt.Errorf("only sys user can register plugin")
	}

	for i, rp := range p.registeredPlugins {
		// 更新插件
		if rp.Name == plugin.Name {
			p.registeredPlugins[i] = plugin
			if err := p.repo.UpdatePlugin(ctx, plugin); err != nil {
				return fmt.Errorf("update plugin error: %v", err)
			}
			log.Infof("update plugin: %v", plugin.String())
			return nil
		}
	}

	// 添加新的代理
	p.registeredPlugins = append(p.registeredPlugins, plugin)
	if err := p.repo.SavePlugin(ctx, plugin); err != nil {
		return fmt.Errorf("add plugin error: %v", err)
	}
	log.Infof("add plugin: %v", plugin.String())
	return nil
}

func (p *PluginUsecase) AddProjectPreCheck(ctx context.Context, ds *Project) error {
	return nil
}

func (p *PluginUsecase) AddProjectAfterHandle(ctx context.Context, ProjectUid string) error {
	if err := p.OperateDataResourceHandle(ctx, ProjectUid, nil, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeCreate, dmsV1.OperationTimingTypeAfter); err != nil {
		return fmt.Errorf("add project handle failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) UpdateProjectPreCheck(ctx context.Context, project *Project) error {
	// 项目归档
	if err := p.OperateDataResourceHandle(ctx, project.UID, dmsV1.IPluginProject{
		Name:     project.Name,
		Archived: project.Status == ProjectStatusArchived,
		Desc:     project.Desc,
	}, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeUpdate, dmsV1.OperationTimingTypeBefore); err != nil {
		return fmt.Errorf("update project handle failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) UpdateProjectAfterHandle(ctx context.Context, projectUid string) error {
	return nil
}

func (p *PluginUsecase) DelProjectPreCheck(ctx context.Context, projectUid string) error {
	if err := p.OperateDataResourceHandle(ctx, projectUid, nil, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore); err != nil {
		return fmt.Errorf("del project pre check failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) DelProjectAfterHandle(ctx context.Context, projectUid string) error {
	if err := p.OperateDataResourceHandle(ctx, projectUid, nil, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeAfter); err != nil {
		return fmt.Errorf("del project handle failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) AddDBServicePreCheck(ctx context.Context, ds *DBService) error {
	dbService := &dmsV1.IPluginDBService{
		Name:             ds.Name,
		DBType:           ds.DBType,
		Host:             ds.Host,
		Port:             ds.Port,
		User:             ds.User,
		Business:         ds.Business,
		AdditionalParams: ds.AdditionalParams,
	}
	if ds.SQLEConfig != nil {
		dbService.SQLERuleTemplateName = ds.SQLEConfig.RuleTemplateName
		dbService.SQLERuleTemplateId = ds.SQLEConfig.RuleTemplateID
	}

	if err := p.OperateDataResourceHandle(ctx, "", dbService, dmsV1.DataResourceTypeDBService, dmsV1.OperationTypeCreate, dmsV1.OperationTimingTypeBefore); err != nil {
		return fmt.Errorf("add db service pre check failed: %v", err)
	}

	return nil
}

func (p *PluginUsecase) AddDBServiceAfterHandle(ctx context.Context, dbServiceUid string) error {
	if err := p.OperateDataResourceHandle(ctx, dbServiceUid, nil, dmsV1.DataResourceTypeDBService, dmsV1.OperationTypeCreate, dmsV1.OperationTimingTypeAfter); err != nil {
		return fmt.Errorf("add db service handle failed: %v", err)
	}

	return nil
}

func (p *PluginUsecase) UpdateDBServicePreCheck(ctx context.Context, ds *DBService) error {
	return nil
}

func (p *PluginUsecase) UpdateDBServiceAfterHandle(ctx context.Context, dbServiceUid string) error {
	if err := p.OperateDataResourceHandle(ctx, dbServiceUid, nil, dmsV1.DataResourceTypeDBService, dmsV1.OperationTypeUpdate, dmsV1.OperationTimingTypeAfter); err != nil {
		return fmt.Errorf("update db service handle failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) DelDBServicePreCheck(ctx context.Context, dbServiceUid string) error {
	if err := p.OperateDataResourceHandle(ctx, dbServiceUid, nil, dmsV1.DataResourceTypeDBService, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore); err != nil {
		return fmt.Errorf("del db service pre check failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) DelDBServiceAfterHandle(ctx context.Context, dbServiceUid string) error {
	if err := p.OperateDataResourceHandle(ctx, dbServiceUid, nil, dmsV1.DataResourceTypeDBService, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeAfter); err != nil {
		return fmt.Errorf("del db service handle failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) DelUserPreCheck(ctx context.Context, userUid string) error {
	if err := p.OperateDataResourceHandle(ctx, userUid, nil, dmsV1.DataResourceTypeUser, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore); err != nil {
		return fmt.Errorf("del user pre check failed: %v", err)
	}
	return nil
}

func (p *PluginUsecase) DelUserGroupPreCheck(ctx context.Context, groupUid string) error {
	return nil
}

func (p *PluginUsecase) OperateDataResourceHandle(ctx context.Context, uid string, resource interface{}, dateResourceType dmsV1.DataResourceType,
	operationType dmsV1.OperationType, operationTiming dmsV1.OperationTimingType) error {
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, plugin := range p.registeredPlugins {
		if plugin.OperateDataResourceHandleUrl != "" {
			wg.Add(1)
			go func(plugin *Plugin) {
				defer wg.Done()
				if err := p.CallOperateDataResourceHandle(ctx, plugin.OperateDataResourceHandleUrl, uid, resource, dateResourceType, operationType, operationTiming); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("call plugin %s operate data resource handle failed: %v", plugin.Name, err))
					mu.Unlock()
				}
			}(plugin)
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("encountered errors: %v", errs)
	}

	return nil
}

func (p *PluginUsecase) CallOperateDataResourceHandle(ctx context.Context, url string, dataResourceUid string, resource interface{}, dataResourceType dmsV1.DataResourceType, operationType dmsV1.OperationType, operationTiming dmsV1.OperationTimingType) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	extraParams, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("marshal resource failed: %v", err)
	}
	operateDataResourceHandleReq := dmsV1.OperateDataResourceHandleReq{
		DataResourceUid:  dataResourceUid,
		DataResourceType: dataResourceType,
		OperationType:    operationType,
		OperationTiming:  operationTiming,
		ExtraParams:      string(extraParams),
	}
	reply := &dmsV1.OperateDataResourceHandleReply{}

	if err := pkgHttp.POST(ctx, url, header, operateDataResourceHandleReq, reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return nil
}
