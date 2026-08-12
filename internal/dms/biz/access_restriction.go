package biz

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

const (
	AccessRestrictionEnabledKey = "access_restriction_enabled"
	AccessPolicyTypeWhitelist   = "whitelist"
)

type AccessWhitelistRule struct {
	Base
	UID        string
	Source     string
	PolicyType string
	Remark     string
}

type AccessRestrictionRepo interface {
	ListRules(ctx context.Context) ([]*AccessWhitelistRule, error)
	GetRuleByUID(ctx context.Context, uid string) (*AccessWhitelistRule, error)
	GetRuleBySource(ctx context.Context, source string) (*AccessWhitelistRule, error)
	CreateRule(ctx context.Context, rule *AccessWhitelistRule) error
	UpdateRule(ctx context.Context, rule *AccessWhitelistRule) error
	DeleteRule(ctx context.Context, uid string) error
	GetEnabled(ctx context.Context) (bool, error)
	SetEnabled(ctx context.Context, enabled bool) error
}

type AccessRestrictionUsecase struct {
	repo AccessRestrictionRepo
	log  *utilLog.Helper
}

func NewAccessRestrictionUsecase(log utilLog.Logger, repo AccessRestrictionRepo) *AccessRestrictionUsecase {
	return &AccessRestrictionUsecase{
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.access_restriction")),
	}
}

func (u *AccessRestrictionUsecase) GetConfig(ctx context.Context) (enabled bool, rules []*AccessWhitelistRule, err error) {
	enabled, err = u.repo.GetEnabled(ctx)
	if err != nil {
		return false, nil, err
	}
	rules, err = u.repo.ListRules(ctx)
	if err != nil {
		return false, nil, err
	}
	return enabled, rules, nil
}

// SetEnabled toggles access restriction. Enabling requires a non-empty whitelist
// and that clientIP matches a rule (same MatchClientIP as future middleware deny).
// Failure does not write enabled=true. Disabling only needs permission (caller).
// Loopback has no privilege: 127.0.0.1 must be explicitly listed.
func (u *AccessRestrictionUsecase) SetEnabled(ctx context.Context, enabled bool, clientIP string) error {
	if !enabled {
		return u.repo.SetEnabled(ctx, false)
	}

	rules, err := u.repo.ListRules(ctx)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return fmt.Errorf("白名单为空，无法开启访问限制")
	}

	matched, err := matchIPAgainstRules(clientIP, rules)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("当前访问来源不在白名单，开启后将无法访问，请先添加当前 IP/网段（检测到的 IP：%s）", clientIP)
	}
	return u.repo.SetEnabled(ctx, true)
}

func (u *AccessRestrictionUsecase) IsEnabled(ctx context.Context) (bool, error) {
	return u.repo.GetEnabled(ctx)
}

func (u *AccessRestrictionUsecase) CreateRule(ctx context.Context, source, remark, policyType string) (*AccessWhitelistRule, error) {
	normalized, err := NormalizeIPv4OrCIDR(source)
	if err != nil {
		return nil, err
	}
	policy, err := normalizePolicyType(policyType)
	if err != nil {
		return nil, err
	}
	exist, err := u.repo.GetRuleBySource(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, fmt.Errorf("来源已存在")
	}
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	rule := &AccessWhitelistRule{
		UID:        uid,
		Source:     normalized,
		PolicyType: policy,
		Remark:     remark,
	}
	if err := u.repo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return u.repo.GetRuleByUID(ctx, uid)
}

func (u *AccessRestrictionUsecase) UpdateRule(ctx context.Context, uid, source, remark, policyType string) (*AccessWhitelistRule, error) {
	existing, err := u.repo.GetRuleByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("规则不存在")
	}
	normalized, err := NormalizeIPv4OrCIDR(source)
	if err != nil {
		return nil, err
	}
	policy, err := normalizePolicyType(policyType)
	if err != nil {
		return nil, err
	}
	conflict, err := u.repo.GetRuleBySource(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if conflict != nil && conflict.UID != uid {
		return nil, fmt.Errorf("来源已存在")
	}
	existing.Source = normalized
	existing.Remark = remark
	existing.PolicyType = policy
	if err := u.repo.UpdateRule(ctx, existing); err != nil {
		return nil, err
	}
	return u.repo.GetRuleByUID(ctx, uid)
}

func (u *AccessRestrictionUsecase) DeleteRule(ctx context.Context, uid string) error {
	existing, err := u.repo.GetRuleByUID(ctx, uid)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("规则不存在")
	}
	return u.repo.DeleteRule(ctx, uid)
}

// MatchClientIP reports whether ip hits any whitelist rule (single IP = /32).
// Shared by enable-guard (S4) and access middleware deny path (S3/AC-004).
func (u *AccessRestrictionUsecase) MatchClientIP(ctx context.Context, ipStr string) (bool, error) {
	rules, err := u.repo.ListRules(ctx)
	if err != nil {
		return false, err
	}
	return matchIPAgainstRules(ipStr, rules)
}

func matchIPAgainstRules(ipStr string, rules []*AccessWhitelistRule) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return false, nil
	}
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if ruleMatchesIP(rule.Source, ip) {
			return true, nil
		}
	}
	return false, nil
}

// ExtractClientIP returns the request source IPv4 from RemoteAddr only (MVP; no XFF trust).
func ExtractClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() == nil {
		return host
	}
	return ip.To4().String()
}

func NormalizeIPv4OrCIDR(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("来源不能为空")
	}
	if strings.Contains(s, ":") {
		return "", fmt.Errorf("不支持 IPv6，请填写 IPv4 或 IPv4 CIDR")
	}
	if strings.Contains(s, "/") {
		_, network, err := net.ParseCIDR(s)
		if err != nil {
			return "", fmt.Errorf("来源格式非法，请填写合法 IPv4 或 IPv4 CIDR")
		}
		if network.IP.To4() == nil {
			return "", fmt.Errorf("不支持 IPv6，请填写 IPv4 或 IPv4 CIDR")
		}
		ones, bits := network.Mask.Size()
		if bits != 32 || ones < 0 || ones > 32 {
			return "", fmt.Errorf("来源格式非法，请填写合法 IPv4 或 IPv4 CIDR")
		}
		return fmt.Sprintf("%s/%d", network.IP.Mask(network.Mask).String(), ones), nil
	}
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("来源格式非法，请填写合法 IPv4 或 IPv4 CIDR")
	}
	return ip.To4().String(), nil
}

func normalizePolicyType(policyType string) (string, error) {
	p := strings.TrimSpace(policyType)
	if p == "" {
		return AccessPolicyTypeWhitelist, nil
	}
	if p != AccessPolicyTypeWhitelist {
		return "", fmt.Errorf("访问策略仅支持白名单")
	}
	return AccessPolicyTypeWhitelist, nil
}

func ruleMatchesIP(source string, ip net.IP) bool {
	if strings.Contains(source, "/") {
		_, network, err := net.ParseCIDR(source)
		if err != nil {
			return false
		}
		return network.Contains(ip)
	}
	ruleIP := net.ParseIP(source)
	if ruleIP == nil {
		return false
	}
	return ruleIP.Equal(ip)
}

// Ensure UpdatedAt is refreshed on update path when storage omits auto-update in map.
func TouchUpdatedAt(rule *AccessWhitelistRule) {
	if rule == nil {
		return
	}
	rule.UpdatedAt = time.Now()
}
