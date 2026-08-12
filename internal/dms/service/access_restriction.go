package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (d *DMSService) GetAccessRestriction(ctx context.Context, currentUserUid string) (*dmsV1.GetAccessRestrictionReply, error) {
	canView, err := d.OpPermissionVerifyUsecase.CanViewGlobal(ctx, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("检查权限失败: %v", err)
	}
	if !canView {
		return nil, fmt.Errorf("无权限查看访问限制配置")
	}

	enabled, rules, err := d.AccessRestrictionUsecase.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &dmsV1.GetAccessRestrictionReply{
		Data: dmsV1.AccessRestrictionConfig{
			Enabled: enabled,
			Rules:   toAccessWhitelistRuleItems(rules),
		},
	}, nil
}

func (d *DMSService) UpdateAccessRestriction(ctx context.Context, currentUserUid string, req *dmsV1.UpdateAccessRestrictionReq, clientIP string) error {
	canOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid, false)
	if err != nil {
		return fmt.Errorf("检查权限失败: %v", err)
	}
	if !canOp {
		return fmt.Errorf("无权限修改访问限制配置")
	}
	if req.Enabled == nil {
		return fmt.Errorf("enabled 不能为空")
	}
	return d.AccessRestrictionUsecase.SetEnabled(ctx, *req.Enabled, clientIP)
}

func (d *DMSService) CreateAccessWhitelistRule(ctx context.Context, currentUserUid string, req *dmsV1.CreateAccessWhitelistRuleReq) (*dmsV1.CreateAccessWhitelistRuleReply, error) {
	canOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid, false)
	if err != nil {
		return nil, fmt.Errorf("检查权限失败: %v", err)
	}
	if !canOp {
		return nil, fmt.Errorf("无权限修改访问限制配置")
	}
	rule, err := d.AccessRestrictionUsecase.CreateRule(ctx, req.Source, req.Remark, req.PolicyType)
	if err != nil {
		return nil, err
	}
	return &dmsV1.CreateAccessWhitelistRuleReply{
		Data: toAccessWhitelistRuleItem(rule),
	}, nil
}

func (d *DMSService) UpdateAccessWhitelistRule(ctx context.Context, currentUserUid string, req *dmsV1.UpdateAccessWhitelistRuleReq) (*dmsV1.UpdateAccessWhitelistRuleReply, error) {
	canOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid, false)
	if err != nil {
		return nil, fmt.Errorf("检查权限失败: %v", err)
	}
	if !canOp {
		return nil, fmt.Errorf("无权限修改访问限制配置")
	}
	rule, err := d.AccessRestrictionUsecase.UpdateRule(ctx, req.RuleUID, req.Source, req.Remark, req.PolicyType)
	if err != nil {
		return nil, err
	}
	return &dmsV1.UpdateAccessWhitelistRuleReply{
		Data: toAccessWhitelistRuleItem(rule),
	}, nil
}

func (d *DMSService) DeleteAccessWhitelistRule(ctx context.Context, currentUserUid string, req *dmsV1.DeleteAccessWhitelistRuleReq) error {
	canOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid, false)
	if err != nil {
		return fmt.Errorf("检查权限失败: %v", err)
	}
	if !canOp {
		return fmt.Errorf("无权限修改访问限制配置")
	}
	return d.AccessRestrictionUsecase.DeleteRule(ctx, req.RuleUID)
}

func (d *DMSService) GetAccessRestrictionClientIP(ctx context.Context, currentUserUid string, r *http.Request) (*dmsV1.GetAccessRestrictionClientIPReply, error) {
	canView, err := d.OpPermissionVerifyUsecase.CanViewGlobal(ctx, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("检查权限失败: %v", err)
	}
	if !canView {
		return nil, fmt.Errorf("无权限查看访问限制配置")
	}
	return &dmsV1.GetAccessRestrictionClientIPReply{
		Data: dmsV1.AccessRestrictionClientIP{
			ClientIP: biz.ExtractClientIP(r),
		},
	}, nil
}

func toAccessWhitelistRuleItems(rules []*biz.AccessWhitelistRule) []dmsV1.AccessWhitelistRuleItem {
	out := make([]dmsV1.AccessWhitelistRuleItem, 0, len(rules))
	for _, rule := range rules {
		out = append(out, toAccessWhitelistRuleItem(rule))
	}
	return out
}

func toAccessWhitelistRuleItem(rule *biz.AccessWhitelistRule) dmsV1.AccessWhitelistRuleItem {
	if rule == nil {
		return dmsV1.AccessWhitelistRuleItem{}
	}
	updatedAt := ""
	if !rule.UpdatedAt.IsZero() {
		updatedAt = rule.UpdatedAt.Format(time.RFC3339)
	}
	return dmsV1.AccessWhitelistRuleItem{
		UID:        rule.UID,
		Source:     rule.Source,
		PolicyType: rule.PolicyType,
		Remark:     rule.Remark,
		UpdatedAt:  updatedAt,
	}
}
