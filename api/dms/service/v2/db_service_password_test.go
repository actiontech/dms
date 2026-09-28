package v2

import (
	"encoding/json"
	"strings"
	"testing"

	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
)

func TestAddDBServiceReq_CipherFieldsPassValidation(t *testing.T) {
	t.Parallel()

	req := &AddDBServiceReq{
		ProjectUid: "700300",
		DBService: &DBService{
			Name:              "mysql_cipher_create",
			DBType:            "MySQL",
			Host:              "10.186.16.126",
			Port:              "3307",
			User:              "testuser",
			SecretPassword:    "cipher-b64",
			EnvironmentTagUID: "2086752861772845056",
			MaintenanceTimes:  nil,
		},
	}
	if err := utilConf.Validate(req); err != nil {
		t.Fatalf("expected cipher create payload to pass Add validation, got: %v", err)
	}
}

func TestAddDBServiceReq_MissingHostStillRequired(t *testing.T) {
	t.Parallel()

	req := &AddDBServiceReq{
		ProjectUid: "700300",
		DBService: &DBService{
			Name:              "mysql_missing_host",
			DBType:            "MySQL",
			Host:              "",
			Port:              "3307",
			User:              "testuser",
			SecretPassword:    "cipher-b64",
			EnvironmentTagUID: "2086752861772845056",
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

func TestDBService_HasPasswordKey(t *testing.T) {
	t.Parallel()

	var withKey DBService
	if err := json.Unmarshal([]byte(`{"name":"n","db_type":"MySQL","host":"127.0.0.1","port":"3306","user":"u","password":"","secret_password":"x","environment_tag_uid":"1"}`), &withKey); err != nil {
		t.Fatal(err)
	}
	if !withKey.HasPasswordKey() {
		t.Fatal("expected password key present")
	}

	var withoutKey DBService
	if err := json.Unmarshal([]byte(`{"name":"n","db_type":"MySQL","host":"127.0.0.1","port":"3306","user":"u","secret_password":"x","environment_tag_uid":"1"}`), &withoutKey); err != nil {
		t.Fatal(err)
	}
	if withoutKey.HasPasswordKey() {
		t.Fatal("expected password key absent")
	}
}

func TestUpdateDBService_HasPasswordKey(t *testing.T) {
	t.Parallel()

	var withKey UpdateDBService
	if err := json.Unmarshal([]byte(`{"db_type":"MySQL","host":"127.0.0.1","port":"3306","user":"u","password":"","secret_password":"x","environment_tag_uid":"1"}`), &withKey); err != nil {
		t.Fatal(err)
	}
	if !withKey.HasPasswordKey() {
		t.Fatal("expected password key present on update")
	}
	if withKey.SecretPassword != "x" {
		t.Fatalf("expected cipher field preserved, got secret=%q", withKey.SecretPassword)
	}
	if !withKey.HasSecretPasswordKey() {
		t.Fatal("expected secret_password key present on update")
	}

	var withoutKey UpdateDBService
	if err := json.Unmarshal([]byte(`{"db_type":"MySQL","host":"127.0.0.1","port":"3306","user":"u","environment_tag_uid":"1","desc":"keep-pwd"}`), &withoutKey); err != nil {
		t.Fatal(err)
	}
	if withoutKey.HasPasswordKey() {
		t.Fatal("expected password key absent on update without password fields")
	}
	if withoutKey.HasSecretPasswordKey() {
		t.Fatal("expected secret_password key absent when omitted")
	}
	if withoutKey.SecretPassword != "" {
		t.Fatal("expected no cipher fields when omitted")
	}

	var emptySecret UpdateDBService
	if err := json.Unmarshal([]byte(`{"db_type":"MySQL","host":"127.0.0.1","port":"3306","user":"u","secret_password":"","environment_tag_uid":"1"}`), &emptySecret); err != nil {
		t.Fatal(err)
	}
	if !emptySecret.HasSecretPasswordKey() {
		t.Fatal("expected empty secret_password placeholder to still set key present")
	}
}
