package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

type nopLogger struct{}

func (nopLogger) Log(_ utilLog.Level, _ ...interface{}) error { return nil }

func newTestDownloadController(enableHttps bool) *DMSController {
	return &DMSController{
		log:         utilLog.NewHelper(nopLogger{}, utilLog.WithMessageKey("test")),
		enableHttps: enableHttps,
	}
}

func TestNodeProxyScheme(t *testing.T) {
	tests := []struct {
		enableHttps bool
		want        string
	}{
		{enableHttps: true, want: "https"},
		{enableHttps: false, want: "http"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := nodeProxyScheme(tt.enableHttps); got != tt.want {
				t.Fatalf("nodeProxyScheme(%v) = %q, want %q", tt.enableHttps, got, tt.want)
			}
		})
	}
}

func serveExportFile(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(echo.HeaderContentDisposition, `attachment; filename="export.zip"`)
		w.Header().Set(echo.HeaderContentType, "application/zip")
		_, _ = io.WriteString(w, body)
	})
}

func invokeProxyDownload(t *testing.T, ctl *DMSController, reportHost string) (*httptest.ResponseRecorder, error) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/dms/projects/p1/data_export_tasks/t1/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := ctl.proxyDownloadDataExportTask(c, reportHost)
	return rec, err
}

func reportHostFromServerURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u.Host
}

func TestProxyDownloadDataExportTaskHTTPSTarget(t *testing.T) {
	const body = "export-zip-bytes"
	server := httptest.NewTLSServer(serveExportFile(body))
	defer server.Close()

	ctl := newTestDownloadController(true)
	rec, err := invokeProxyDownload(t, ctl, reportHostFromServerURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != body {
		t.Fatalf("body = %q, want %q", got, body)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderContentDisposition), "export.zip") {
		t.Fatalf("missing content-disposition, got %q", rec.Header().Get(echo.HeaderContentDisposition))
	}
}

func TestProxyDownloadDataExportTaskHTTPTarget(t *testing.T) {
	const body = "export-zip-http"
	server := httptest.NewServer(serveExportFile(body))
	defer server.Close()

	ctl := newTestDownloadController(false)
	rec, err := invokeProxyDownload(t, ctl, reportHostFromServerURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != body {
		t.Fatalf("body = %q, want %q", got, body)
	}
}

func TestProxyDownloadDataExportTaskSchemeMismatch(t *testing.T) {
	const body = "should-not-receive"
	server := httptest.NewTLSServer(serveExportFile(body))
	defer server.Close()

	// Old bug: always used http against an HTTPS peer.
	ctl := newTestDownloadController(false)
	rec, err := invokeProxyDownload(t, ctl, reportHostFromServerURL(server.URL))
	if err == nil && rec.Code == http.StatusOK && rec.Body.String() == body {
		t.Fatal("expected scheme mismatch to fail download, but got successful response")
	}
	if err != nil {
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("error type = %T, want *echo.HTTPError", err)
		}
		if httpErr.Code != http.StatusBadGateway {
			t.Fatalf("error code = %d, want %d", httpErr.Code, http.StatusBadGateway)
		}
		return
	}
	// Plain HTTP against a Go TLS listener typically yields HTTP/1.0 400.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d or transport error; body=%q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestProxyDownloadDataExportTaskUnreachable(t *testing.T) {
	server := httptest.NewServer(serveExportFile("unused"))
	reportHost := reportHostFromServerURL(server.URL)
	server.Close()

	ctl := newTestDownloadController(false)
	_, err := invokeProxyDownload(t, ctl, reportHost)
	if err == nil {
		t.Fatal("expected error for unreachable target")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("error type = %T, want *echo.HTTPError", err)
	}
	if httpErr.Code != http.StatusBadGateway {
		t.Fatalf("error code = %d, want %d", httpErr.Code, http.StatusBadGateway)
	}
	msg := fmt.Sprint(httpErr.Message)
	if !strings.Contains(msg, "could not forward") {
		t.Fatalf("error message = %q, want contain %q", msg, "could not forward")
	}
}
