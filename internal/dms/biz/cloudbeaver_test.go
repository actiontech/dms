package biz

import (
	"testing"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// Test_GenerateCloudbeaverConnectionParams_PolarDB covers design.md §11.2 Table 4
// (plan 5.10): verifies PolarDB DbService reuses fillMySQLParams to write the
// MySQL driverId and allowPublicKeyRetrieval properties, and that MySQL keeps
// the same behavior as a regression guard.
func Test_GenerateCloudbeaverConnectionParams_PolarDB(t *testing.T) {
	cu := &CloudbeaverUsecase{
		log: utilLog.NewHelper(&noopLogger{}, utilLog.WithMessageKey("test")),
	}

	cases := map[string]struct {
		dbType                   string
		wantDriverID             string
		wantAllowPublicKeyRetVal string
	}{
		// (a) PolarDB happy: driverId == "mysql:mysql8" (plan 3.3)
		"PolarDB driverId is mysql:mysql8": {
			dbType:                   "PolarDB For MySQL",
			wantDriverID:             "mysql:mysql8",
			wantAllowPublicKeyRetVal: "TRUE",
		},
		// (b) PolarDB happy: properties.allowPublicKeyRetrieval == "TRUE" (fillMySQLParams reuse)
		"PolarDB allowPublicKeyRetrieval is TRUE": {
			dbType:                   "PolarDB For MySQL",
			wantDriverID:             "mysql:mysql8",
			wantAllowPublicKeyRetVal: "TRUE",
		},
		// (c) MySQL regression: confirm existing MySQL behavior is unchanged
		"MySQL regression keeps driverId mysql:mysql8": {
			dbType:                   "MySQL",
			wantDriverID:             "mysql:mysql8",
			wantAllowPublicKeyRetVal: "TRUE",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbService := &DBService{
				Name:     "test-instance",
				DBType:   tc.dbType,
				Host:     "127.0.0.1",
				Port:     "3306",
				User:     "root",
				Password: "pwd",
			}

			resp, err := cu.GenerateCloudbeaverConnectionParams(dbService, "proj", "")
			if err != nil {
				t.Fatalf("GenerateCloudbeaverConnectionParams err: %v", err)
			}

			config, ok := resp["config"].(map[string]interface{})
			if !ok {
				t.Fatalf("config is not map[string]interface{}: %T", resp["config"])
			}

			if got := config["driverId"]; got != tc.wantDriverID {
				t.Errorf("driverId = %v, want %v", got, tc.wantDriverID)
			}

			props, ok := config["properties"].(map[string]interface{})
			if !ok {
				t.Fatalf("properties is not map[string]interface{}: %T", config["properties"])
			}
			if got := props["allowPublicKeyRetrieval"]; got != tc.wantAllowPublicKeyRetVal {
				t.Errorf("allowPublicKeyRetrieval = %v, want %v", got, tc.wantAllowPublicKeyRetVal)
			}
		})
	}
}
