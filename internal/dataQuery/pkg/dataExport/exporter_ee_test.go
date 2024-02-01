package data_export

import (
	"fmt"
	"os"
	"testing"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

func TestExport(t *testing.T) {
	db, err := NewMysqlConn("10.186.56.59", "33063", "root", "123", "dms")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	logger := utilLog.NewHelper(utilLog.NewMyLogger(os.Stdout), utilLog.WithMessageKey("storage.role"))
	result, err := ExportTasksToZip(logger, "./test.zip", []*ExportTask{
		NewExportTask().WithExtract(NewExtract(db, "select name, email, phone from users")).WithExporter("users_1.csv", NewCsvExport()),
		NewExportTask().WithExtract(NewExtract(db, "select name, email, phone from ttttttt limit 1")).WithExporter("users_2.csv", NewCsvExport()),
	})
	t.Log(result)
	if err != nil {
		return
	}
}
