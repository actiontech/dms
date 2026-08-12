package biz

import (
	"net/url"
	"testing"

	"github.com/labstack/echo/v4/middleware"
)

func TestIsRegisteredServiceIP(t *testing.T) {
	u1, _ := url.Parse("http://10.1.2.3:5432")
	u2, _ := url.Parse("http://odc.example.com:8989") // domain → no exempt
	d := &DmsProxyUsecase{
		targets: []*ProxyTarget{
			{ProxyTarget: middleware.ProxyTarget{Name: "sqle", URL: u1}},
			{ProxyTarget: middleware.ProxyTarget{Name: "odc-dns", URL: u2}},
		},
	}
	if !d.IsRegisteredServiceIP("10.1.2.3") {
		t.Fatal("registered IP should exempt")
	}
	if d.IsRegisteredServiceIP("127.0.0.1") {
		t.Fatal("unregistered loopback must not auto-exempt")
	}
	if d.IsRegisteredServiceIP("odc.example.com") {
		t.Fatal("hostname string is not client IP")
	}
}
