package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/actiontech/dms/internal/pkg/locale"
	pkgPeriods "github.com/actiontech/dms/pkg/periods"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// MaintenanceTimeConfig 运维时间管控配置
type MaintenanceTimeConfig struct {
	Enabled bool
	Periods pkgPeriods.Periods
}

// MaintenanceTimeUsecase 运维时间管控业务逻辑
type MaintenanceTimeUsecase struct {
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	log                       *utilLog.Helper
}

// NewMaintenanceTimeUsecase 创建运维时间管控业务逻辑实例
func NewMaintenanceTimeUsecase(
	log utilLog.Logger,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
) *MaintenanceTimeUsecase {
	return &MaintenanceTimeUsecase{
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.maintenance_time")),
	}
}

// MaintenanceTimeConfigFromSQLQuery 从数据源 sql_query_config 解析运维时间管控配置
func MaintenanceTimeConfigFromSQLQuery(sqlQueryConfig *SQLQueryConfig) *MaintenanceTimeConfig {
	cfg := &MaintenanceTimeConfig{}
	if sqlQueryConfig == nil {
		return cfg
	}
	cfg.Enabled = len(sqlQueryConfig.MaintenancePeriods) > 0
	if cfg.Enabled {
		cfg.Periods = sqlQueryConfig.MaintenancePeriods.Copy()
	}
	return cfg
}

// CheckSQLExecutionAllowed 检查SQL执行是否被运维时间管控允许
// 参数:
//   - userUid: 当前执行SQL的用户UID
//   - sqlTypes: SQLE审核返回的sql_type列表（每条SQL对应一个）
//   - currentTime: 当前时间（参数化便于测试）
//   - sqlQueryConfig: 数据源上的 SQL 查询配置（含运维时间窗口）
//
// 返回值:
//   - allowed: 是否允许执行
//   - message: 拦截时的提示消息（allowed=true时为空）
//   - err: 内部错误
func (m *MaintenanceTimeUsecase) CheckSQLExecutionAllowed(
	ctx context.Context,
	userUid string,
	sqlTypes []string,
	currentTime time.Time,
	sqlQueryConfig *SQLQueryConfig,
) (allowed bool, message string, err error) {
	// 1. 获取配置
	config := MaintenanceTimeConfigFromSQLQuery(sqlQueryConfig)

	// 2. 如果开关关闭，直接允许执行
	if !config.Enabled {
		return true, "", nil
	}

	// 3. 检查 sqlTypes 中是否有非DQL语句
	hasNonDQL := false
	for _, sqlType := range sqlTypes {
		if sqlType != "dql" { // 空字符串("")也视为非DQL（保守策略）
			hasNonDQL = true
			break
		}
	}
	if !hasNonDQL {
		return true, "", nil
	}

	// 4. 检查用户是否为管理员
	isAdmin, err := m.opPermissionVerifyUsecase.CanOpGlobal(ctx, userUid)
	if err != nil {
		return false, "", fmt.Errorf("failed to check user admin permission: %v", err)
	}
	if isAdmin {
		m.log.Warnf("user %s is admin, skip cloudbeaver maintenance time check", userUid)
		return true, "", nil
	}

	// 5. 检查当前时间是否在配置的运维时间段内
	if config.Periods.IsWithinScope(currentTime) {
		return true, "", nil
	}

	// 6. 构造拦截消息
	periodsStr := formatPeriodsToReadableString(config.Periods)
	message = fmt.Sprintf(locale.Bundle.LocalizeMsgByCtx(ctx, locale.SqlWorkbenchMaintenanceTimeBlocked), periodsStr)

	// 7. 返回拦截结果
	return false, message, nil
}

// formatPeriodsToReadableString 将时间段格式化为可读字符串
// 例如: "01:00-06:00, 22:00-02:00"
func formatPeriodsToReadableString(ps pkgPeriods.Periods) string {
	parts := make([]string, 0, len(ps))
	for _, p := range ps {
		parts = append(parts, fmt.Sprintf("%02d:%02d-%02d:%02d",
			p.StartHour, p.StartMinute, p.EndHour, p.EndMinute))
	}
	return strings.Join(parts, ", ")
}
