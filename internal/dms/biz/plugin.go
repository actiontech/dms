package biz

import (
	"context"
	"fmt"

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
	AddDBServicePreCheckUrl      string
	DelDBServicePreCheckUrl      string
	DelUserPreCheckUrl           string
	DelUserGroupPreCheckUrl      string
	OperateDataResourceHandleUrl string
}

func (p *Plugin) String() string {
	return fmt.Sprintf("name=%v,addDBServicePreCheckUrl=%v,delDBServicePreCheckUrl=%v,delUserPreCheckUrl=%v,delUserGroupPreCheckUrl=%v,OperateDataHandleUrl=%v",
		p.Name, p.AddDBServicePreCheckUrl, p.DelDBServicePreCheckUrl, p.DelUserPreCheckUrl, p.DelUserGroupPreCheckUrl, p.OperateDataResourceHandleUrl)
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

func (p *PluginUsecase) AddDBServicePreCheck(ctx context.Context, ds *DBService) error {
	dbTyp, err := dmsV1.ParseIPluginDBType(ds.DBType.String())
	if err != nil {
		return fmt.Errorf("parse db type failed: %v", err)
	}
	dbService := &dmsV1.IPluginDBService{
		Name:     ds.Name,
		DBType:   dbTyp,
		Host:     ds.Host,
		Port:     ds.Port,
		User:     ds.User,
		Business: ds.Business,
	}
	if ds.SQLEConfig != nil {
		dbService.SQLERuleTemplateName = ds.SQLEConfig.RuleTemplateName
		dbService.SQLERuleTemplateId = ds.SQLEConfig.RuleTemplateID
	}
	for _, plugin := range p.registeredPlugins {
		if plugin.AddDBServicePreCheckUrl != "" {
			if err := p.CallAddDBServicePreCheck(ctx, plugin.AddDBServicePreCheckUrl, dbService); err != nil {
				return fmt.Errorf("plugin %s add db service pre check failed: %v", plugin.Name, err)
			}
		}
	}
	return nil
}

func (p *PluginUsecase) DelDBServicePreCheck(ctx context.Context, dbServiceUid string) error {
	for _, plugin := range p.registeredPlugins {
		if plugin.DelDBServicePreCheckUrl != "" {
			if err := p.CallDelDBServicePreCheck(ctx, plugin.DelDBServicePreCheckUrl, dbServiceUid); err != nil {
				return fmt.Errorf("plugin %s del db service pre check failed: %v", plugin.Name, err)
			}
		}
	}
	return nil
}

func (p *PluginUsecase) DelUserPreCheck(ctx context.Context, userUid string) error {
	for _, plugin := range p.registeredPlugins {
		if plugin.DelUserPreCheckUrl != "" {
			if err := p.CallDelUserPreCheck(ctx, plugin.DelUserPreCheckUrl, userUid); err != nil {
				return fmt.Errorf("plugin %s del user pre check failed: %v", plugin.Name, err)
			}
		}
	}
	return nil
}

func (p *PluginUsecase) DelUserGroupPreCheck(ctx context.Context, groupUid string) error {
	for _, plugin := range p.registeredPlugins {
		if plugin.DelUserGroupPreCheckUrl != "" {
			if err := p.CallDelUserGroupPreCheck(ctx, plugin.DelUserGroupPreCheckUrl, groupUid); err != nil {
				return fmt.Errorf("plugin %s del user group pre check failed: %v", plugin.Name, err)
			}
		}
	}
	return nil
}

func (p *PluginUsecase) OperateDataResourceHandle(ctx context.Context, uid string, dateResourceType dmsV1.DataResourceType,
	operationType dmsV1.OperationType, operationTiming dmsV1.OperationTimingType) error {
	for _, plugin := range p.registeredPlugins {
		if plugin.OperateDataResourceHandleUrl != "" {
			if err := p.CallOperateDataResourceHandle(ctx, plugin.OperateDataResourceHandleUrl, uid, dateResourceType, operationType, operationTiming); err != nil {
				return fmt.Errorf("call plugin %s operate data resource handle failed: %v", plugin.Name, err)
			}
		}
	}
	return nil
}

func (p *PluginUsecase) CallAddDBServicePreCheck(ctx context.Context, url string, ds *dmsV1.IPluginDBService) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reqBody := struct {
		DBService *dmsV1.IPluginDBService `json:"db_service"`
	}{
		DBService: ds,
	}

	reply := &dmsV1.AddDBServicePreCheckReply{}

	if err := pkgHttp.Get(ctx, url, header, reqBody, reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return nil
}

func (p *PluginUsecase) CallDelDBServicePreCheck(ctx context.Context, url string, dbServiceUid string) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reqBody := struct {
		DBServiceUid string `json:"db_service_uid"`
	}{
		DBServiceUid: dbServiceUid,
	}

	reply := &dmsV1.DelDBServicePreCheckReply{}

	if err := pkgHttp.Get(ctx, url, header, reqBody, reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return nil
}

func (p *PluginUsecase) CallDelUserPreCheck(ctx context.Context, url string, userUid string) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reqBody := struct {
		UserUid string `json:"user_uid"`
	}{
		UserUid: userUid,
	}

	reply := &dmsV1.DelUserPreCheckReply{}

	if err := pkgHttp.Get(ctx, url, header, reqBody, reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return nil
}

func (p *PluginUsecase) CallDelUserGroupPreCheck(ctx context.Context, url string, userGroupUid string) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reqBody := struct {
		UserGroupUid string `json:"user_group_uid"`
	}{
		UserGroupUid: userGroupUid,
	}
	reply := &dmsV1.DelUserGroupPreCheckReply{}

	if err := pkgHttp.Get(ctx, url, header, reqBody, reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return nil
}

func (p *PluginUsecase) CallOperateDataResourceHandle(ctx context.Context, url string, dataResourceUid string, dataResourceType dmsV1.DataResourceType, operationType dmsV1.OperationType, operationTiming dmsV1.OperationTimingType) error {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	operateDataResourceHandleReq := dmsV1.OperateDataResourceHandleReq{
		DataResourceUid:  dataResourceUid,
		DataResourceType: dataResourceType,
		OperationType:    operationType,
		OperationTiming:  operationTiming,
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
