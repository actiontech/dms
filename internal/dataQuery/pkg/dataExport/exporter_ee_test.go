package data_export

import (
	"fmt"
	"testing"
)

func TestExport(t *testing.T) {
	db, err := NewMysqlConn("10.186.62.16", "3306", "root", "mysqlpass", "dms")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	err = ExportTasksToZip("./test.zip", []*ExportTask{
		NewExportTask().WithExtract(NewExtract(db, "select name, email, phone from users")).WithExporter("users_1.csv", NewCsvExport()),
		NewExportTask().WithExtract(NewExtract(db, "select name, email, phone from users limit 1")).WithExporter("users_2.csv", NewCsvExport()),
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}
