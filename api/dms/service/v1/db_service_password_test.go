package v1

import (
	"strings"
	"testing"

	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
)

func TestAddDBServiceReq_EmptyPasswordRejected(t *testing.T) {
	t.Parallel()

	base := func(password string) *AddDBServiceReq {
		return &AddDBServiceReq{
			ProjectUid: "700300",
			DBService: &DBService{
				Name:             "gbase8a_empty_pwd",
				DBType:           "GBase-8a",
				Host:             "10.186.16.126",
				Port:             "5258",
				User:             "root",
				Password:         password,
				Business:         "default",
				MaintenanceTimes: nil,
			},
		}
	}

	t.Run("password_empty_string", func(t *testing.T) {
		t.Parallel()
		err := utilConf.Validate(base(""))
		if err == nil {
			t.Fatal("expected empty password to fail Add validation")
		}
		msg := strings.ToLower(err.Error())
		if !strings.Contains(msg, "password") || !strings.Contains(msg, "required") {
			t.Fatalf("expected Password required validation error, got: %v", err)
		}
	})

	t.Run("password_non_empty_passes_password_rule", func(t *testing.T) {
		t.Parallel()
		if err := utilConf.Validate(base("not-empty")); err != nil {
			t.Fatalf("expected non-empty password to pass Add validation, got: %v", err)
		}
	})
}

func TestAddDBServiceReq_MissingHostStillRequired(t *testing.T) {
	t.Parallel()

	req := &AddDBServiceReq{
		ProjectUid: "700300",
		DBService: &DBService{
			Name:     "gbase8a_missing_host",
			DBType:   "GBase-8a",
			Host:     "",
			Port:     "5258",
			User:     "root",
			Password: "not-empty",
			Business: "default",
		},
	}

	err := utilConf.Validate(req)
	if err == nil {
		t.Fatal("expected missing Host to fail validation")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Host") || !strings.Contains(msg, "required") {
		t.Fatalf("expected Host required validation error, got: %v", err)
	}
}
