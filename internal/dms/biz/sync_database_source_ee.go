//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"strings"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
)

const (
	databaseSourceDMPSupportedVersion = "5.23.01.0"
	dmpToken                          = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJoYXNoIjoiWzUyIDMgMTI5IDE4NCAyMzggNjcgMiAzMCA5NSAyMzkgNTMgMjExIDE1MyAxNzQgODkgMjQ3XSIsIm5iZiI6MTY3MDU2MTg1NSwic2VlZCI6ImdaQ0J4RFJxUm1DQzJpaG4iLCJ1c2VyIjoic3FsZSJ9.SxGEi6QP8Dtl3ChsetDeZQxbcYpqsXmQibmytRuDbsg"
)

type DatabaseSourceImpl interface {
	SyncDatabaseSource(context.Context, *DatabaseSourceServiceParams, *DatabaseSourceServiceUsecase, string) error
}

var databaseSourceMap = map[pkgConst.DBServiceSourceName]DatabaseSourceImpl{}

func GetDatabaseSourceImpl(name pkgConst.DBServiceSourceName) (DatabaseSourceImpl, error) {
	databaseSourceImpl, ok := databaseSourceMap[name]
	if ok {
		return databaseSourceImpl, nil
	}

	return nil, fmt.Errorf("%s hasn't implemented", name)
}

func init() {
	databaseSourceMap[pkgConst.DBServiceSourceNameDMP] = dmpManager{}
}

type dmpManager struct {
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

func (d dmpManager) SyncDatabaseSource(ctx context.Context, params *DatabaseSourceServiceParams, serviceUsecase *DatabaseSourceServiceUsecase, currentUserId string) error {
	if params.Version < databaseSourceDMPSupportedVersion {
		return fmt.Errorf("dmp version %s not supported", params.Version)
	}

	dmpFilterType := d.getDmpFilterType(params.DbType)

	url := fmt.Sprintf("%s/v3/support/data_sources?filter_by_type=%s", params.URL, dmpFilterType)

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
			Field:    string(DatabaseSourceServiceFieldNamespaceUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    params.NamespaceUID,
		},
	}

	dbServices, err := serviceUsecase.dbServiceUsecase.repo.GetDBServices(ctx, conditions)
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
			serviceUsecase.log.Errorf("dmp data source %s sip is empty", item.DataSrcID)
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
			serviceUsecase.namespaceUsecase.log.Warnf("can not get business from remote: %s, port %s", item.DataSrcSip, item.DataSrcPort)
			continue
		}

		sourceId := item.DataSrcID

		desc := fmt.Sprintf("sync dmp database source: %v", item.Tags)
		dbServiceParams := &BizDBServiceArgs{
			Name:          item.DataSrcID,
			Desc:          &desc,
			DBType:        params.DbType,
			Host:          item.DataSrcSip,
			Port:          item.DataSrcPort,
			AdminUser:     item.DataSrcUser,
			AdminPassword: &password,
			Business:      strings.Join(businessArr, ","),
			NamespaceUID:  params.NamespaceUID,
			Source:        string(pkgConst.DBServiceSourceNameDMP),
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
			_, err = serviceUsecase.dbServiceUsecase.CreateDBService(ctx, dbServiceParams, currentUserId)
		} else {
			remainDBServiceSourceMap[sourceId] = dbService.UID
			// update
			if dbService.Host != item.DataSrcSip || dbService.Port != item.DataSrcPort || dbService.AdminUser != item.DataSrcUser || dbService.AdminPassword != password {
				err = serviceUsecase.dbServiceUsecase.UpdateDBService(ctx, dbService.UID, dbServiceParams, currentUserId)
			}
		}

		if err != nil {
			return err
		}
	}

	for sourceId, dbService := range dbServiceSourceAddrMap {
		if _, ok := remainDBServiceSourceMap[sourceId]; !ok {
			// delete db service
			if err = serviceUsecase.dbServiceUsecase.DelDBService(ctx, dbService.UID, currentUserId); err != nil {
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
