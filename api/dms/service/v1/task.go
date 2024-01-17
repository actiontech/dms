package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:parameters AddDataExportTask
type AddDataExportTaskReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// add data export workflow
	// in:body
	DataExportWorkflow []DataExportTask `json:"data_export_workflow"`
}

type DataExportTask struct {
	// DB Service uid
	// Required: true
	DBServiceUid string `json:"db_service_uid" validate:"required"`
	// DB Service name
	// Required: false
	DatabaseName string `json:"database_name"`
	// The exported SQL statement executed. it's necessary when ExportType is SQL
	// Required: true
	// SELECT * FROM DMS_test LIMIT 20;
	ExecSQL string `json:"exec_sql" validate:"required"`

	// // Export Type
	// // Required: false
	// // example: SQL
	// ExportType string `json:"export_type" validate:"required"`
	// // Export Content
	// // Required: true
	// // example: DATA STRUCTURE DATA_AND_STRUCTUR
	// ExportContent string `json:"export_content" validate:"required"`
	// // Export Content
	// // Required: true
	// // example: CSV SQL EXCEL
	// ExportFileType string `json:"export_file_type" validate:"required"`
	// Export Content
	// Required: true
	// example:  UTF-8 GBK
	// ExportCharset string `json:"export_charset" validate:"required"`
	// // The exported tables. it's necessary when ExportType is DATABASE
	// // Required: false
	// //  [{name: "users", schemaName: "dms", all: "custom", columns: ["id", "age"]}]
	// Tables []Table `json:"tables"`
	// // The exported tables. it's necessary when ExportType is Database
	// // Required: false
	// // ["BINARY" ,"TEXT" , "BLOB"]
	// DataOption []string `json:"data_option"`
	// // The exported sql option. it's necessary when ExportContent is Database
	// // Required: false
	// // ["COMPRESS", "TRUNCATE", "DROP"]
	// SQLOption []string `json:"sql_option"`
	// // The exported sql option. it's necessary when ExportContent is Database
	// // Required: false
	// // ["FUNCTION", "VIEW", "EVENT", "PROCEDURE", "TRIGGER"]
	// StructureOption []string `json:"structure_option"`
}

// type Table struct {
// 	Name       string   `json:"name"`
// 	SchemaName string   `json:"schemaName"`
// 	All        string   `json:"all"`
// 	Columns    []string `json:"columns"`
// 	Condition  string   `json:"condition"`
// }

// swagger:model AddDataExportTaskReply
type AddDataExportTaskReply struct {
	// add data export workflow reply
	Data struct {
		// data export task UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}
