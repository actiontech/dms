package constant

import "fmt"

// internel build-in uid
const (
	UIDOfOpPermissionCreateProject       = "700001"
	UIDOfOpPermissionProjectAdmin        = "700002"
	UIDOfOpPermissionCreateWorkflow      = "700003"
	UIDOfOpPermissionAuditWorkflow       = "700004"
	UIDOfOpPermissionAuthDBServiceData   = "700005"
	UIDOfOpPermissionExecuteWorkflow     = "700006"
	UIDOfOpPermissionViewOthersWorkflow  = "700007"
	UIDOfOpPermissionViewOthersAuditPlan = "700008"
	UIDOfOpPermissionSaveAuditPlan       = "700009"
	UIDOfOpPermissionSQLQuery            = "700010"

	UIDOfDMSConfig = "700100"

	UIDOfUserAdmin = "700200"
	UIDOfUserSys   = "700201"

	UIDOfProjectDefault = "700300"

	UIDOfRoleProjectAdmin   = "700400"
	UIDOfRoleSQLEAdmin      = "700401"
	UIDOfRoleProvisionAdmin = "700402"
)

type DBType string

func (d DBType) String() string {
	return string(d)
}

func ParseDBType(s string) (DBType, error) {
	switch s {
	case "MySQL":
		return DBTypeMySQL, nil
	case "PostgreSQL":
		return DBTypePostgreSQL, nil
	case "Oracle":
		return DBTypeOracle, nil
	case "SQLServer":
		return DBTypeSQLServer, nil
	case "OceanBaseMySQL":
		return DBTypeOceanBaseMySQL, nil
	default:
		return "", fmt.Errorf("invalid db type: %s", s)
	}
}

const (
	DBTypeMySQL          DBType = "MySQL"
	DBTypePostgreSQL     DBType = "PostgreSQL"
	DBTypeOracle         DBType = "Oracle"
	DBTypeSQLServer      DBType = "SQLServer"
	DBTypeOceanBaseMySQL DBType = "OceanBaseMySQL"
)

type FilterCondition struct {
	KeywordSearch bool
	Field         string
	Operator      FilterOperator
	Value         interface{}
}

type FilterOperator string

const (
	FilterOperatorEqual              FilterOperator = "="
	FilterOperatorIsNull             FilterOperator = "isNull"
	FilterOperatorNotEqual           FilterOperator = "<>"
	FilterOperatorContains           FilterOperator = "like"
	FilterOperatorGreaterThanOrEqual FilterOperator = ">="
	FilterOperatorLessThanOrEqual    FilterOperator = "<="
	FilterOperatorIn                 FilterOperator = "in"
)

type DBServiceSourceName string

const (
	DBServiceSourceNameDMP  DBServiceSourceName = "Actiontech DMP"
	DBServiceSourceNameDMS  DBServiceSourceName = "Actiontech DMS"
	DBServiceSourceNameSQLE DBServiceSourceName = "SQLE"
)

func ParseDBServiceSource(s string) (DBServiceSourceName, error) {
	switch s {
	case string(DBServiceSourceNameDMP):
		return DBServiceSourceNameDMP, nil
	case string(DBServiceSourceNameDMS):
		return DBServiceSourceNameDMS, nil
	case string(DBServiceSourceNameSQLE):
		return DBServiceSourceNameSQLE, nil
	default:
		return "", fmt.Errorf("invalid data object source name: %s", s)
	}
}

const (
	DMSToken = "dms-token"
)
