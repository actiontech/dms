package middleware

import "testing"

func TestNormalizeRequestPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/v1/dms/projects/1746123456789012345/db_services", "/v1/dms/projects/:uid/db_services"},
		{"/sqle/v2/projects/demo/workflows/123/tasks/456", "/sqle/v2/projects/demo/workflows/:id/tasks/:id"},
		{"/sql_query/api/gql?foo=bar", "/sql_query/api/gql"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := normalizeRequestPath(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeRequestPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsTrackableUserActivityPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/assets/index-abc123.js", false},
		{"/static/logo.png", false},
		{"/favicon.ico", false},
		{"/user-activity", false},
		{"/v1/dms/personalization/logo", false},
		{"/v1/dms/statistic/user_activity/summary", false},
		{"/v1/dms/projects", true},
		{"/sqle/v1/projects/demo/workflows", true},
		{"/sql_query/api/gql", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTrackableUserActivityPath(tt.path)
			if got != tt.want {
				t.Fatalf("isTrackableUserActivityPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestResolveModuleCode(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/sqle/v2/projects/:name/workflows", "WORKFLOW"},
		{"/sql_query/api/gql", "WORKBENCH"},
		{"/v1/dms/operation_records", "AUDIT_LOG"},
		{"/unknown/path", "OTHER"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := resolveModuleCode(tt.path)
			if got != tt.want {
				t.Fatalf("resolveModuleCode(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
