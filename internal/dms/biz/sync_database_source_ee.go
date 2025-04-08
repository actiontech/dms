//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgError "github.com/actiontech/dms/internal/dms/pkg/errors"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
)

const (
	databaseSourceDMPSupportedVersion = "5.23.01.0"
	dmpToken                          = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJoYXNoIjoiWzUyIDMgMTI5IDE4NCAyMzggNjcgMiAzMCA5NSAyMzkgNTMgMjExIDE1MyAxNzQgODkgMjQ3XSIsIm5iZiI6MTY3MDU2MTg1NSwic2VlZCI6ImdaQ0J4RFJxUm1DQzJpaG4iLCJ1c2VyIjoic3FsZSJ9.SxGEi6QP8Dtl3ChsetDeZQxbcYpqsXmQibmytRuDbsg"
)

type DatabaseSourceImpl interface {
	SyncDatabaseSource(context.Context, *DBServiceSyncTask, string) error
}

func NewDatabaseSourceImpl(name pkgConst.DBServiceSourceName, syncTaskUsecase *DBServiceSyncTaskUsecase) (DatabaseSourceImpl, error) {
	switch name {
	case pkgConst.DBServiceSourceNameDMP:
		return dmpManager{
			syncTaskUsecase: syncTaskUsecase,
		}, nil
	case pkgConst.DBServiceSourceNameExpandService:
		return expandService{
			syncTaskUsecase: syncTaskUsecase,
		}, nil
	}
	return nil, fmt.Errorf("%s hasn't implemented", name)
}

type dmpManager struct {
	syncTaskUsecase *DBServiceSyncTaskUsecase
}

type Tag struct {
	// tag attribute. 标签名
	TagAttribute string `json:"tag_attribute"`
	// tag value. 标签值
	TagValue string `json:"tag_value"`
}

type ListService struct {
	// data source id. 数据源ID, 例如实例组 ID
	DataSrcID string `json:"data_src_id,omitempty"`

	// data source encrypted password. 数据源密码（已加密）
	DataSrcPassword string `json:"data_src_password,omitempty"`

	// data source port. 数据源端口
	DataSrcPort string `json:"data_src_port,omitempty"`

	// data source ip. 数据源的SIP，即实例组的SIP，或者固定不变的实例IP
	DataSrcSip string `json:"data_src_sip,omitempty"`

	// user name. 数据源用户名
	DataSrcUser string `json:"data_src_user,omitempty"`

	// tags  业务标签.
	Tags []Tag `json:"tags,omitempty"`
}

type DmpDatabaseSourceResp struct {
	// data
	Data []*ListService `json:"data"`

	// total number of data sources. 数据源列表总计
	TotalNums uint32 `json:"total_nums"`
}

func (d dmpManager) SyncDatabaseSource(ctx context.Context, params *DBServiceSyncTask, currentUserId string) error {
	dmpVersion := params.AdditionalParam.GetParam("version").String()
	if dmpVersion < databaseSourceDMPSupportedVersion {
		return fmt.Errorf("dmp version %s not supported", dmpVersion)
	}

	dbType, err := pkgConst.ParseDBType(params.DbType)
	if err != nil {
		return err
	}
	dmpFilterType := d.getDmpFilterType(dbType)

	url := fmt.Sprintf("%s/v3/support/data_sources?page_index=1&page_size=9999&filter_by_type=%s", params.URL, dmpFilterType)

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": dmpToken,
	}
	var resp = &DmpDatabaseSourceResp{}

	if err := pkgHttp.Get(ctx, url, header, nil, resp); err != nil {
		return err
	}

	if resp.TotalNums < 1 {
		return fmt.Errorf("dmp data source total nums %d less than 1", resp.TotalNums)
	}

	conditions := []pkgConst.FilterCondition{
		{
			Field:    string(DatabaseSourceServiceFieldSource),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    params.Source,
		},
		{
			Field:    string(DatabaseSourceServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    pkgConst.UIDOfProjectDefault,
		},
	}

	dbServices, err := d.syncTaskUsecase.dbServiceUsecase.repo.GetDBServices(ctx, conditions)
	if err != nil {
		return err
	}

	dbServiceSourceAddrMap := make(map[string]*DBService)
	for _, dbService := range dbServices {
		dbServiceSourceAddrMap[dbService.Name] = dbService
	}

	remainDBServiceSourceMap := make(map[string]string)
	for _, item := range resp.Data {
		password, err := DecryptPassword(item.DataSrcPassword)
		if err != nil {
			return fmt.Errorf("sync dmp database source decrypt password err: %v", err)
		}

		if item.DataSrcSip == "" {
			d.syncTaskUsecase.log.Errorf("dmp data source %s sip is empty", item.DataSrcID)
			continue
		}

		businessArr := make([]string, 0)
		for _, tag := range item.Tags {
			if strings.Contains(tag.TagAttribute, "业务") || strings.Contains(tag.TagAttribute, "business") {
				businessArr = append(businessArr, tag.TagValue)
			}
		}
		// 业务为空不支持同步
		if len(businessArr) == 0 {
			d.syncTaskUsecase.projectUsecase.log.Warnf("can not get business from remote: %s, port %s", item.DataSrcSip, item.DataSrcPort)
			continue
		}

		sourceId := item.DataSrcID

		desc := fmt.Sprintf("sync dmp database source: %v", item.Tags)
		dbServiceParams := &BizDBServiceArgs{
			Name:       item.DataSrcID,
			Desc:       &desc,
			DBType:     params.DbType,
			Host:       item.DataSrcSip,
			Port:       item.DataSrcPort,
			User:       item.DataSrcUser,
			Password:   &password,
			// Business:   strings.Join(businessArr, ","),
			ProjectUID: pkgConst.UIDOfProjectDefault,
			Source:     string(pkgConst.DBServiceSourceNameDMP),
		}

		sqleConfig := params.SQLEConfig
		if sqleConfig != nil {
			dbServiceParams.RuleTemplateName = sqleConfig.RuleTemplateName
			dbServiceParams.RuleTemplateID = sqleConfig.RuleTemplateID
			if sqleConfig.SQLQueryConfig != nil {
				dbServiceParams.SQLQueryConfig = &SQLQueryConfig{
					MaxPreQueryRows:                  sqleConfig.SQLQueryConfig.MaxPreQueryRows,
					QueryTimeoutSecond:               sqleConfig.SQLQueryConfig.QueryTimeoutSecond,
					AuditEnabled:                     sqleConfig.SQLQueryConfig.AuditEnabled,
					AllowQueryWhenLessThanAuditLevel: sqleConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel,
				}
			}
		}

		dbService, ok := dbServiceSourceAddrMap[sourceId]
		if !ok {
			// create
			_, err = d.syncTaskUsecase.dbServiceUsecase.CreateDBService(ctx, dbServiceParams, currentUserId)
		} else {
			remainDBServiceSourceMap[sourceId] = dbService.UID
			// update
			if dbService.Host != item.DataSrcSip || dbService.Port != item.DataSrcPort || dbService.User != item.DataSrcUser || dbService.Password != password {
				err = d.syncTaskUsecase.dbServiceUsecase.UpdateDBServiceByArgs(ctx, dbService.UID, dbServiceParams, currentUserId)
			}
		}

		if err != nil {
			return err
		}
	}

	for sourceId, dbService := range dbServiceSourceAddrMap {
		if _, ok := remainDBServiceSourceMap[sourceId]; !ok {
			// delete db service
			if err = d.syncTaskUsecase.dbServiceUsecase.DelDBService(ctx, dbService.UID, currentUserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d dmpManager) getDmpFilterType(dbType pkgConst.DBType) string {
	switch dbType {
	case pkgConst.DBTypeMySQL:
		return "mysql"
	case pkgConst.DBTypePostgreSQL:
		return ""
	case pkgConst.DBTypeOracle:
		return ""
	case pkgConst.DBTypeSQLServer:
		return ""
	case pkgConst.DBTypeOceanBaseMySQL:
		return ""
	default:
		return ""
	}
}

type expandService struct {
	syncTaskUsecase *DBServiceSyncTaskUsecase
}

type SyncDBServiceResV1 struct {
	DBServices []SyncDBService `json:"db_services"`
}

type SyncDBService struct {
	DBService
	ProjectName string `json:"project_name"`
}

func (s expandService) SyncDatabaseSource(ctx context.Context, syncTask *DBServiceSyncTask, currentUserId string) error {
	var header = map[string]string{"Content-Type": "application/json"}
	var resp = &SyncDBServiceResV1{}

	if err := pkgHttp.Get(ctx, syncTask.URL, header, nil, resp); err != nil {
		return fmt.Errorf("when sync database source, get data from %v failed, err:%v", syncTask.URL, err)
	}
	// map db services by project name
	projectSyncDBServiceMap := make(map[string] /*project name*/ []*SyncDBService, len(resp.DBServices))
	for _, dbService := range resp.DBServices {
		db := dbService
		db.Source = syncTask.Name
		projectSyncDBServiceMap[dbService.ProjectName] = append(projectSyncDBServiceMap[dbService.ProjectName], &db)
	}
	// update db services by project
	for projectName, dbServices := range projectSyncDBServiceMap {
		project, err := getOrCreateProject(ctx, projectName, fmt.Sprintf("project sync by external database sync task, named: %v", syncTask.Name), s.syncTaskUsecase)
		if err != nil {
			return fmt.Errorf("when sync database source, getOrCreateProject failed, err:%v", err)
		}
		if err := s.syncDBServices(ctx, dbServices, project, syncTask, currentUserId); err != nil {
			return fmt.Errorf("when sync database source, syncDBServices, err:%v", err)
		}
	}
	return nil
}

func getOrCreateProject(ctx context.Context, projectName, projectDesc string, syncTaskUsecase *DBServiceSyncTaskUsecase) (*Project, error) {
	project, err := syncTaskUsecase.projectUsecase.GetProjectByName(ctx, projectName)
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, pkgError.ErrStorageNoData) {
		return nil, err
	}
	// TODO 批量创建项目目前不支持配置项目优先级，先按照中配置
	project, err = NewProject(pkgConst.UIDOfUserAdmin, projectName, projectDesc, dmsCommonV1.ProjectPriorityMedium, project.BusinessTag.UID)
	if err != nil {
		return nil, err
	}
	if err = syncTaskUsecase.projectUsecase.CreateProject(ctx, project, pkgConst.UIDOfUserAdmin); err != nil {
		return nil, err
	}
	return syncTaskUsecase.projectUsecase.GetProjectByName(ctx, projectName)
}

func (s expandService) syncDBServices(ctx context.Context, syncDBServices []*SyncDBService, project *Project, syncTask *DBServiceSyncTask, currentUserId string) error {
	// get db services from project
	currentDBServices, err := s.syncTaskUsecase.dbServiceUsecase.repo.GetDBServices(ctx, []pkgConst.FilterCondition{
		{
			Field:    string(DatabaseSourceServiceFieldSource),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    syncTask.Name,
		},
		{
			Field:    string(DatabaseSourceServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    project.UID,
		},
	})
	if err != nil {
		return fmt.Errorf("when sync db service get db service in project failed, project uid: %v, db type: %v, err: %v", project.UID, syncTask.DbType, err)
	}
	// map current db service by name
	currentDBServiceMap := make(map[string] /*db service name*/ *DBService, len(currentDBServices))
	for _, currentDBService := range currentDBServices {
		db := currentDBService
		currentDBServiceMap[currentDBService.Name] = db
	}

	for _, dbService := range syncDBServices {
		dbServiceParams := convertDbServiceToDbParams(dbService, project)
		if db, exist := currentDBServiceMap[dbServiceParams.Name]; exist {
			// if exist update db service
			err = s.syncTaskUsecase.dbServiceUsecase.UpdateDBServiceByArgs(ctx, db.UID, &dbServiceParams, currentUserId)
			if err != nil {
				return err
			}
			// delete updated db service from map
			delete(currentDBServiceMap, dbServiceParams.Name)
		} else {
			// if not exist create db service
			_, err := s.syncTaskUsecase.dbServiceUsecase.CreateDBService(ctx, setDefaultSQLEConfig(&dbServiceParams, syncTask), currentUserId)
			if err != nil {
				return err
			}
		}

	}
	// delete remaining databases
	for _, dbService := range currentDBServiceMap {
		return s.syncTaskUsecase.dbServiceUsecase.DelDBService(ctx, dbService.UID, currentUserId)
	}
	return nil
}

func convertDbServiceToDbParams(dbService *SyncDBService, project *Project) BizDBServiceArgs {
	dbServiceParams := BizDBServiceArgs{
		Name:              dbService.Name,
		Desc:              &dbService.Desc,
		DBType:            dbService.DBType,
		Host:              dbService.Host,
		Port:              dbService.Port,
		User:              dbService.User,
		Password:          &dbService.Password,
		// Business:          dbService.Business,
		Source:            dbService.Source,
		AdditionalParams:  dbService.AdditionalParams,
		ProjectUID:        project.UID,
		MaintenancePeriod: dbService.MaintenancePeriod,
		IsMaskingSwitch:   dbService.IsMaskingSwitch,
	}

	if dbService.SQLEConfig != nil {
		dbServiceParams.RuleTemplateName = dbService.SQLEConfig.RuleTemplateName
		dbServiceParams.RuleTemplateID = dbService.SQLEConfig.RuleTemplateID
		dbServiceParams.SQLQueryConfig = dbService.SQLEConfig.SQLQueryConfig
	}
	return dbServiceParams
}

func setDefaultSQLEConfig(dbServiceParams *BizDBServiceArgs, syncTask *DBServiceSyncTask) *BizDBServiceArgs {
	if syncTask.SQLEConfig == nil {
		return dbServiceParams
	}
	if dbServiceParams.RuleTemplateID == "" {
		dbServiceParams.RuleTemplateID = syncTask.SQLEConfig.RuleTemplateID
		dbServiceParams.RuleTemplateName = syncTask.SQLEConfig.RuleTemplateName
	}
	if dbServiceParams.SQLQueryConfig == nil {
		dbServiceParams.SQLQueryConfig = &SQLQueryConfig{
			AllowQueryWhenLessThanAuditLevel: syncTask.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel,
			AuditEnabled:                     syncTask.SQLEConfig.SQLQueryConfig.AuditEnabled,
			MaxPreQueryRows:                  syncTask.SQLEConfig.SQLQueryConfig.MaxPreQueryRows,
			QueryTimeoutSecond:               syncTask.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond,
		}
	}
	return dbServiceParams
}
