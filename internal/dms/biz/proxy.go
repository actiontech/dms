package biz

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ProxyTargetRepo interface {
	SaveProxyTarget(ctx context.Context, u *ProxyTarget) error
	UpdateProxyTarget(ctx context.Context, u *ProxyTarget) error
	ListProxyTargets(ctx context.Context) ([]*ProxyTarget, error)
	ListProxyTargetsByScenarios(ctx context.Context, scenarios []ProxyScenario) ([]*ProxyTarget, error)
	GetProxyTargetByName(ctx context.Context, name string) (*ProxyTarget, error)
}

type ProxyTarget struct {
	middleware.ProxyTarget
	Version  string
	Scenario ProxyScenario
}
type ProxyScenario string

const (
	ProxyScenarioInternalService     ProxyScenario = "internal_service"
	ProxyScenarioThirdPartyIntegrate ProxyScenario = "thrid_party_integrate"
	ProxyScenarioUnknown             ProxyScenario = "unknown"
)

var ErrUnknownProxyScenario error = fmt.Errorf("unknown proxy scenairo")

const ProxyTargetMetaKey = "prefixs"

func (p *ProxyTarget) GetProxyUrlPrefixs() []string {
	ret, _ := p.Meta[ProxyTargetMetaKey].([]string)
	return ret
}

func (p *ProxyTarget) SetProxyUrlPrefix(prefixs []string) {
	p.Meta[ProxyTargetMetaKey] = prefixs
}

type DmsProxyUsecase struct {
	repo              ProxyTargetRepo
	targets           []*ProxyTarget
	defaultTargetSelf *ProxyTarget
	rewrite           map[string]string
	mutex             sync.RWMutex
	logger            utilLog.Logger
	opPermissionUc    *OpPermissionUsecase
}

func NewDmsProxyUsecase(logger utilLog.Logger, repo ProxyTargetRepo, dmsPort int, opPermissionUC *OpPermissionUsecase) (*DmsProxyUsecase, error) {
	targets, err := repo.ListProxyTargets(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("list proxy targets from repo error: %v", err)
	}

	dmsUrl, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%v", dmsPort))
	return &DmsProxyUsecase{
		repo: repo,
		// 将自身定义为默认代理，当无法匹配转发规则时，转发到自身
		defaultTargetSelf: &ProxyTarget{
			ProxyTarget: middleware.ProxyTarget{
				Name: _const.DmsComponentName,
				URL:  dmsUrl,
			},
		},
		// TODO 支持可配置
		rewrite: map[string]string{
			"/sqle/*":    "/$1",
			"/webhook/*": "/$1",
		},
		targets:        targets,
		logger:         logger,
		opPermissionUc: opPermissionUC,
	}, nil
}

type RegisterDMSProxyTargetArgs struct {
	Name            string
	Addr            string
	Version         string
	ProxyUrlPrefixs []string
	Scenario        ProxyScenario
}

func (d *DmsProxyUsecase) GetTargetByName(ctx context.Context, name string) (*ProxyTarget, error) {
	return d.repo.GetProxyTargetByName(ctx, name)
}

func (d *DmsProxyUsecase) RegisterDMSProxyTarget(ctx context.Context, currentUserUid string, args RegisterDMSProxyTargetArgs) error {
	log := utilLog.NewHelper(d.logger, utilLog.WithMessageKey("biz.dmsproxy"))

	// check if the user is sys
	if currentUserUid != pkgConst.UIDOfUserSys {
		return fmt.Errorf("only sys user can register proxy")
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if err := d.checkProxyUrlPrefix(args.ProxyUrlPrefixs); err != nil {
		return err
	}

	url, err := url.ParseRequestURI(args.Addr)
	if err != nil {
		return fmt.Errorf("invalid url: %s", args.Addr)
	}

	target := &ProxyTarget{
		ProxyTarget: middleware.ProxyTarget{
			Name: args.Name,
			URL:  url,
			Meta: echo.Map{ProxyTargetMetaKey: args.ProxyUrlPrefixs},
		},
		Version:  args.Version,
		Scenario: args.Scenario,
	}

	for i, t := range d.targets {
		// 更新代理
		if t.Name == target.Name {
			d.targets[i] = target
			if err := d.repo.UpdateProxyTarget(ctx, target); err != nil {
				return fmt.Errorf("update proxy target error: %v", err)
			}
			log.Infof("update target: %s; url: %s; prefix: %v", target.Name, target.URL, args.ProxyUrlPrefixs)
			return nil
		}
	}

	// 添加新的代理
	d.targets = append(d.targets, target)
	if err := d.repo.SaveProxyTarget(ctx, target); err != nil {
		return fmt.Errorf("add proxy target error: %v", err)
	}
	log.Infof("add target: %s; url: %s; prefix: %v", target.Name, target.URL, args.ProxyUrlPrefixs)

	// 注册独立权限
	if err := d.opPermissionUc.InitOpPermissions(ctx, GetProxyOpPermission()[target.Name]); nil != err {
		return err
	}

	return nil
}

func (d *DmsProxyUsecase) checkProxyUrlPrefix(proxyUrlPrefixs []string) error {
	for _, prefix := range proxyUrlPrefixs {
		for _, t := range d.targets {
			if t.Meta[ProxyTargetMetaKey] == prefix {
				return fmt.Errorf("proxy url prefix: %s already exists", prefix)
			}
		}
	}

	return nil
}

func (d *DmsProxyUsecase) ListProxyTargets(ctx context.Context) ([]*ProxyTarget, error) {
	return d.repo.ListProxyTargets(ctx)
}

func (d *DmsProxyUsecase) ListProxyTargetsByScenarios(ctx context.Context, scenarios []ProxyScenario) ([]*ProxyTarget, error) {
	return d.repo.ListProxyTargetsByScenarios(ctx, scenarios)
}

// AddTarget实现echo的ProxyBalancer接口， 没有实际意义
func (d *DmsProxyUsecase) AddTarget(target *middleware.ProxyTarget) bool {
	return true
}

// RemoveTarget 实现echo的ProxyBalancer接口，没有实际意义
func (d *DmsProxyUsecase) RemoveTarget(name string) bool {
	return true
}

// Next 实现echo的ProxyBalancer接口，定义转发逻辑，echo会使用该转发逻辑进行转发
func (d *DmsProxyUsecase) Next(c echo.Context) *middleware.ProxyTarget {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log := utilLog.NewHelper(d.logger, utilLog.WithMessageKey("biz.dmsproxy.Next"))

	for _, t := range d.targets {
		for _, prefix := range t.GetProxyUrlPrefixs() {
			if prefix != "" && strings.HasPrefix(c.Request().URL.Path, prefix) {
				log.Debugf("url: %s; proxy to target: %s; proxy prefix: %v", c.Request().URL.Path, t.Name, t.Meta[ProxyTargetMetaKey])
				return &middleware.ProxyTarget{
					Name: t.Name,
					URL:  t.URL,
					Meta: t.Meta,
				}
			}
		}
	}

	// 由于Skipper方法的存在，当无法匹配转发规则时，会跳过转发，所以大部分情况下不会执行到这里。
	// 极端情况比如Skipper后，target列表发生了变动，则可能执行到这里，使用默认代理转发到自身作为兜底。
	log.Debugf("proxy to default target")

	return &middleware.ProxyTarget{
		Name: d.defaultTargetSelf.Name,
		URL:  d.defaultTargetSelf.URL,
		Meta: d.defaultTargetSelf.Meta,
	}
}

// DmsProxyUsecase 实现了 middleware.ProxyBalancer 接口
func (d *DmsProxyUsecase) GetEchoProxyBalancer() middleware.ProxyBalancer {
	return d
}

func (d *DmsProxyUsecase) GetEchoProxySkipper() middleware.Skipper {
	return d.Skipper
}

func (d *DmsProxyUsecase) GetEchoProxyRewrite() map[string]string {
	return d.rewrite
}

// 当无法匹配转发规则时，跳过转发
func (d *DmsProxyUsecase) Skipper(c echo.Context) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, t := range d.targets {
		for _, prefix := range t.GetProxyUrlPrefixs() {
			if prefix != "" && strings.HasPrefix(c.Request().URL.Path, prefix) {
				return false
			}
		}
	}

	return true
}
