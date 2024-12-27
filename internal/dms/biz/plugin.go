package biz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
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
	Name string
	// 该地址目的是统一调用其他服务 数据资源变更前后校验/更新数据的 接口
	// eg: 删除数据源前：
	// 需要sqle服务中实现接口逻辑，判断该数据源上已经没有进行中的工单
	OperateDataResourceHandleUrl string
	GetDatabaseDriverOptionsUrl  string
	GetDatabaseDriverLogosUrl    string
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

const (
	LogoPath = "/logo/"
	LogoDir  = "./static/logo/"
)

var databaseDriverOptions []*v1.DatabaseDriverOption

func (p *PluginUsecase) GetDatabaseDriverOptionsCache() []*v1.DatabaseDriverOption {
	return databaseDriverOptions
}

func (p *PluginUsecase) ClearDatabaseDriverOptionsCache() {
	databaseDriverOptions = []*v1.DatabaseDriverOption{}
}

func (p *PluginUsecase) GetDatabaseDriverOptionsHandle(ctx context.Context) ([]*v1.DatabaseDriverOption, error) {
	log := utilLog.NewHelper(p.logger, utilLog.WithMessageKey("biz.dmsplugin.DatabaseDriverOptionsHandle"))
	cacheOptions := p.GetDatabaseDriverOptionsCache()
	if len(cacheOptions) != 0 {
		return cacheOptions, nil
	}
	var (
		mu        sync.Mutex
		errs      []error
		wg        sync.WaitGroup
		dbOptions []struct {
			options []*v1.DatabaseDriverOption
			source  string
		}
	)

	for _, plugin := range p.registeredPlugins {
		if plugin.GetDatabaseDriverOptionsUrl == "" {
			continue
		}
		wg.Add(1)
		go func(plugin *Plugin) {
			defer wg.Done()
			op, err := p.CallDatabaseDriverOptionsHandle(ctx, plugin.GetDatabaseDriverOptionsUrl)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("call plugin %s get database driver options handle failed: %v", plugin.Name, err))
				mu.Unlock()
				return
			}
			mu.Lock()
			dbOptions = append(dbOptions, struct {
				options []*v1.DatabaseDriverOption
				source  string
			}{
				options: op,
				source:  plugin.Name,
			})
			mu.Unlock()
		}(plugin)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("encountered errors: %v", errs)
	}
	options := p.aggregateOptions(log, dbOptions)
	databaseDriverOptions = append(databaseDriverOptions, options...)
	dbTypes := make([]string, 0, len(databaseDriverOptions))
	for _, dbType := range databaseDriverOptions {
		dbTypes = append(dbTypes, dbType.DBType)
	}
	// 处理数据库插件logo
	go p.DatabaseLogoHandle(context.TODO(), dbTypes)
	return options, nil
}

// 根据数据库类型合并各插件的options
func (p *PluginUsecase) aggregateOptions(log *utilLog.Helper, optionRes []struct {
	options []*v1.DatabaseDriverOption
	source  string
}) []*v1.DatabaseDriverOption {
	dbTypeMap := make(map[string]*v1.DatabaseDriverOption)
	for _, res := range optionRes {
		for _, opt := range res.options {
			if aggOpt, exists := dbTypeMap[opt.DBType]; exists {
				// 聚合Params, 合并时如有重复以sqle为主
				aggOpt.Params = mergeParamsByName(aggOpt.Params, opt.Params, res.source == _const.SqleComponentName)
			} else {
				logofile, err := p.GetLogoFilesMap(opt.DBType)
				if err != nil {
					log.Errorf("get %s logo file name error: %v", opt.DBType, err)
				}
				dbTypeMap[opt.DBType] = &v1.DatabaseDriverOption{
					DBType:   opt.DBType,
					LogoPath: LogoPath + logofile[opt.DBType],
					Params:   opt.Params,
				}
			}
		}
	}

	// 转换为切片返回
	result := make([]*v1.DatabaseDriverOption, 0, len(dbTypeMap))
	for _, opt := range dbTypeMap {
		result = append(result, opt)
	}
	return result
}

// 根据参数名合并additional和params, overwriteExisting代表是不是要以新参数覆盖旧参数
func mergeParamsByName(existing, newParams []*v1.DatabaseDriverAdditionalParam, overwriteExisting bool) []*v1.DatabaseDriverAdditionalParam {
	paramMap := make(map[string]*v1.DatabaseDriverAdditionalParam)

	// 添加已有参数
	for _, param := range existing {
		paramMap[param.Name] = param
	}

	// 合并新参数
	for _, param := range newParams {
		if _, exists := paramMap[param.Name]; exists && overwriteExisting {
			newAggParam := *param
			paramMap[param.Name] = &newAggParam // 覆盖已有参数
		} else if !exists {
			paramMap[param.Name] = param
		}
	}

	// 转换为切片返回
	result := make([]*v1.DatabaseDriverAdditionalParam, 0, len(paramMap))
	for _, param := range paramMap {
		result = append(result, param)
	}
	return result
}

func (p *PluginUsecase) CallDatabaseDriverOptionsHandle(ctx context.Context, url string) ([]*v1.DatabaseDriverOption, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	reply := &v1.ListDBServiceDriverOptionReply{}

	if err := pkgHttp.Get(ctx, url, header, nil, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}

func (p *PluginUsecase) DatabaseLogoHandle(ctx context.Context, dbTypes []string) {
	log := utilLog.NewHelper(p.logger, utilLog.WithMessageKey("biz.dmsplugin.logohandle"))
	// 定义 logo 文件夹路径
	if err := os.MkdirAll(LogoDir, os.ModePerm); err != nil {
		log.Errorf("crate logo dir error: %v", err)
		return
	}
	// 获取需要保存logo的数据库插件类型
	logoDBTypes, err := p.GetDBTypesForLogoSave(dbTypes)
	if err != nil {
		log.Errorf("get db types for logo save error: %v", err)
		return
	}
	if len(logoDBTypes) == 0 {
		return
	}
	var (
		mu       sync.Mutex
		errs     []error
		wg       sync.WaitGroup
		allLogos []struct {
			logo   string
			dbType string
			source string
		}
	)
	for _, plugin := range p.registeredPlugins {
		if plugin.GetDatabaseDriverLogosUrl == "" {
			continue
		}

		wg.Add(1)
		go func(plugin *Plugin) {
			defer wg.Done()
			logos, err := p.CallDatabaseDriverLogosHandle(ctx, plugin.GetDatabaseDriverLogosUrl, logoDBTypes)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("call plugin %s get database driver logos handle failed: %v", plugin.Name, err))
				mu.Unlock()
				return
			}
			mu.Lock()
			for k, v := range logos {
				allLogos = append(allLogos, struct {
					logo   string
					dbType string
					source string
				}{
					logo:   v,
					dbType: k,
					source: plugin.Name,
				})
			}
			mu.Unlock()
		}(plugin)
	}
	wg.Wait()

	if len(errs) > 0 {
		log.Errorf("encountered errors: %v", errs)
	}
	logoMap := make(map[string]string)
	for _, entry := range allLogos {
		// 目前只使用sqle提供的logo，因为其他插件暂未提供logo
		if _, found := logoMap[entry.dbType]; !found && entry.source == _const.SqleComponentName {
			logoMap[entry.dbType] = entry.logo
		}
	}
	err = p.SaveLogoFiles(log, logoMap)
	if err != nil {
		log.Errorf("save logo error: %v", err)
	}
}

func (p *PluginUsecase) GetDBTypesForLogoSave(allDBTypes []string) ([]string, error) {
	needToSave := []string{}
	existingFiles, err := p.GetLogoFilesMap(allDBTypes...)
	if err != nil {
		return nil, fmt.Errorf("failed to read logo directory '%s': %w", LogoDir, err)
	}
	// 检查每种数据库类型对应的文件是否已存在
	for _, dbType := range allDBTypes {
		// 如果未找到对应文件，加入需要保存的列表
		if _, ok := existingFiles[dbType]; !ok {
			needToSave = append(needToSave, dbType)
		}
	}

	return needToSave, nil
}

// 返回指定数据库类型的logo文件名称
func (p *PluginUsecase) GetLogoFilesMap(dbTypes ...string) (map[string]string /*key: db type, value: logo file name*/, error) {
	entries, err := os.ReadDir(LogoDir)
	if err != nil {
		return nil, err
	}
	logoFiles := make(map[string]string)
	for _, dbType := range dbTypes {
		// 构建文件名前缀
		filePrefix := strings.ToLower(strings.ReplaceAll(dbType, " ", "_")) + "."
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), filePrefix) {
				logoFiles[dbType] = entry.Name()
			}
		}
	}
	return logoFiles, nil
}

func (p *PluginUsecase) CallDatabaseDriverLogosHandle(ctx context.Context, url string, dbTypes []string) (map[string]string /*key: db type; value: logo string*/, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	var reply struct {
		Logos []struct {
			DBType string `json:"db_type"`
			Logo   string `json:"logo"`
		} `json:"data"`
		base.GenericResp
	}
	reqBody := struct {
		DBTypes string `json:"db_types"`
	}{
		DBTypes: strings.Join(dbTypes, ","),
	}
	if err := pkgHttp.Get(ctx, url, header, reqBody, &reply); err != nil {
		return nil, fmt.Errorf("failed to get logos for %s: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}
	logosMap := make(map[string]string, len(reply.Logos))
	for _, logo := range reply.Logos {
		logosMap[logo.DBType] = logo.Logo
	}
	return logosMap, nil
}

func (p *PluginUsecase) SaveLogoFiles(log *utilLog.Helper, logoMap map[string]string /*key: db type; value: logo string*/) error {
	for k, v := range logoMap {
		if v == "" {
			log.Errorf("%s logo base64 string is empty", k)
			continue
		}
		data, err := base64.StdEncoding.DecodeString(v)

		if err != nil {
			log.Errorf("decode %s logo base64 string error: %v", k, err)
			continue
		}
		mimeType := http.DetectContentType(data)
		isSupport, logoType := p.GetLogoFileTypeByHttpContentType(mimeType)
		if !isSupport {
			log.Errorf("unsupported image type: %s", mimeType)
			continue
		}
		fileName := p.GetLogoFileName(k, logoType)
		if err := os.WriteFile(filepath.Join(LogoDir, fileName), data, os.ModePerm); err != nil {
			log.Errorf("write %s logo file error: %v", k, err)
			continue
		}
	}
	p.ClearDatabaseDriverOptionsCache()
	return nil
}

func (p *PluginUsecase) GetLogoFileName(dbType, logoType string) string {
	return strings.ToLower(strings.ReplaceAll(dbType, " ", "_")) + logoType
}

func (p *PluginUsecase) GetLogoFileTypeByHttpContentType(mimeType string) (bool, string) {
	switch mimeType {
	case "image/jpeg":
		return true, ".jpeg"
	case "image/png":
		return true, ".png"
	case "image/svg+xml":
		return true, ".svg"
	case "image/webp":
		return true, ".webp"
	default:
		return false, ""
	}
}
