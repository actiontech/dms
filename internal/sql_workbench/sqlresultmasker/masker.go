package sqlresultmasker

import "context"

// MaskWorkbenchResultsArgs carries tabular SQL workbench (ODC) result data and masking scope. Rows are masked in place.
type MaskWorkbenchResultsArgs struct {
	Rows         [][]any  `json:"rows"`
	ColumnNames  []string `json:"column_names"`
	SQL          string   `json:"sql"`
	DBServiceUID string   `json:"db_service_uid"`
	SchemaName   string   `json:"schema_name"`
	ProjectUID   string   `json:"project_uid"`
}

// SQLResultMasker masks SQL workbench (ODC) tabular result rows in place and reports which columns were masked.
type SQLResultMasker interface {
	MaskSQLWorkbenchResults(ctx context.Context, args *MaskWorkbenchResultsArgs) (map[string]bool, error)
}
