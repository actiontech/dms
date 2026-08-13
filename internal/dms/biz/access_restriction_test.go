package biz

import (
	"context"
	"io"
	"strings"
	"testing"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

func TestNormalizeIPv4OrCIDR(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"192.168.1.1", "192.168.1.1", false},
		{"10.0.0.0/24", "10.0.0.0/24", false},
		{"10.0.0.8/24", "10.0.0.0/24", false},
		{"", "", true},
		{"not-an-ip", "", true},
		{"2001:db8::1", "", true},
		{"1.2.3.4/33", "", true},
	}
	for _, c := range cases {
		got, err := NormalizeIPv4OrCIDR(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("NormalizeIPv4OrCIDR(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeIPv4OrCIDR(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("NormalizeIPv4OrCIDR(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizePolicyType(t *testing.T) {
	got, err := normalizePolicyType("")
	if err != nil || got != AccessPolicyTypeWhitelist {
		t.Fatalf("empty policy: got=%q err=%v", got, err)
	}
	if _, err := normalizePolicyType("blacklist"); err == nil {
		t.Fatal("blacklist should be rejected")
	}
}

func TestMatchIPAgainstRules(t *testing.T) {
	rules := []*AccessWhitelistRule{
		{Source: "192.168.1.100"},
		{Source: "10.8.0.0/24"},
	}
	ok, err := matchIPAgainstRules("10.8.0.5", rules)
	if err != nil || !ok {
		t.Fatalf("CIDR hit: ok=%v err=%v", ok, err)
	}
	ok, err = matchIPAgainstRules("127.0.0.1", rules)
	if err != nil || ok {
		t.Fatalf("loopback not listed: ok=%v err=%v", ok, err)
	}
	ok, err = matchIPAgainstRules("not-ip", rules)
	if err != nil || ok {
		t.Fatalf("invalid ip: ok=%v err=%v", ok, err)
	}
}

type memAccessRestrictionRepo struct {
	enabled bool
	rules   []*AccessWhitelistRule
}

func (m *memAccessRestrictionRepo) ListRules(ctx context.Context) ([]*AccessWhitelistRule, error) {
	return m.rules, nil
}
func (m *memAccessRestrictionRepo) GetRuleByUID(ctx context.Context, uid string) (*AccessWhitelistRule, error) {
	return nil, nil
}
func (m *memAccessRestrictionRepo) GetRuleBySource(ctx context.Context, source string) (*AccessWhitelistRule, error) {
	return nil, nil
}
func (m *memAccessRestrictionRepo) CreateRule(ctx context.Context, rule *AccessWhitelistRule) error {
	return nil
}
func (m *memAccessRestrictionRepo) UpdateRule(ctx context.Context, rule *AccessWhitelistRule) error {
	return nil
}
func (m *memAccessRestrictionRepo) DeleteRule(ctx context.Context, uid string) error { return nil }
func (m *memAccessRestrictionRepo) GetEnabled(ctx context.Context) (bool, error) {
	return m.enabled, nil
}
func (m *memAccessRestrictionRepo) SetEnabled(ctx context.Context, enabled bool) error {
	m.enabled = enabled
	return nil
}

func TestSetEnabledGuard(t *testing.T) {
	repo := &memAccessRestrictionRepo{enabled: false}
	u := NewAccessRestrictionUsecase(utilLog.NewMyLogger(io.Discard), repo)

	if err := u.SetEnabled(context.Background(), true, "127.0.0.1"); err == nil || !strings.Contains(err.Error(), "白名单为空") {
		t.Fatalf("empty list enable: err=%v", err)
	}
	if repo.enabled {
		t.Fatal("empty list must keep switch off")
	}

	repo.rules = []*AccessWhitelistRule{{Source: "192.168.1.100"}}
	if err := u.SetEnabled(context.Background(), true, "127.0.0.1"); err == nil || !strings.Contains(err.Error(), "检测到的 IP：127.0.0.1") {
		t.Fatalf("miss enable: err=%v", err)
	}
	if repo.enabled {
		t.Fatal("miss must keep switch off")
	}

	repo.rules = []*AccessWhitelistRule{{Source: "127.0.0.1"}}
	if err := u.SetEnabled(context.Background(), true, "127.0.0.1"); err != nil {
		t.Fatalf("hit enable: %v", err)
	}
	if !repo.enabled {
		t.Fatal("hit should enable")
	}
	if err := u.SetEnabled(context.Background(), false, "10.0.0.1"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if repo.enabled {
		t.Fatal("disable should turn off")
	}
}
