package service

import (
	"context"
	"errors"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

	"github.com/go-openapi/strfmt"
)

func (d *DMSService) ListNamespaces(ctx context.Context, req *dmsCommonV1.ListProjectReq, currentUserUid string) (reply *dmsCommonV1.ListProjectReply, err error) {
	var orderBy biz.NamespaceField
	switch req.OrderBy {
	case dmsCommonV1.ProjectOrderByName:
		orderBy = biz.NamespaceFieldName
	default:
		orderBy = biz.NamespaceFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.NamespaceFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}
	if req.FilterByUID != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.NamespaceFieldUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByUID,
		})
	}

	listOption := &biz.ListNamespacesOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	namespaces, total, err := d.NamespaceUsecase.ListNamespace(ctx, listOption, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsCommonV1.ListProject, len(namespaces))
	for i, n := range namespaces {
		ret[i] = &dmsCommonV1.ListProject{
			ProjectUid: n.UID,
			Name:       n.Name,
			Archived:   (n.Status == biz.NamespaceStatusArchived),
			Desc:       n.Desc,
			CreateTime: strfmt.DateTime(n.CreateTime),
		}
		user, err := d.UserUsecase.GetUser(ctx, n.CreateUserUID)
		if err != nil {
			d.log.Errorf("get user error: %v", err)
			continue
		}
		ret[i].CreateUser = dmsCommonV1.UidWithName{
			Uid:  n.UID,
			Name: user.Name,
		}

	}

	return &dmsCommonV1.ListProjectReply{
		Payload: struct {
			Projects []*dmsCommonV1.ListProject `json:"projects"`
			Total    int64                      `json:"total"`
		}{Projects: ret, Total: total},
	}, nil
}

func (d *DMSService) AddNamespace(ctx context.Context, currentUserUid string, req *dmsV1.AddProjectReq) (reply *dmsV1.AddProjectReply, err error) {
	// check
	{
		// check current user has enough permission
		if canCreateNamespace, err := d.OpPermissionVerifyUsecase.CanCreateNamespace(ctx, currentUserUid); err != nil {
			return nil, err
		} else if !canCreateNamespace {
			return nil, fmt.Errorf("current user can't create namespace")
		}

		// check project is exist
		_, err := d.NamespaceUsecase.GetNamespaceByName(ctx, req.Project.Name)
		if err == nil {
			return nil, fmt.Errorf("namespace %v is exist", req.Project.Name)
		}
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, fmt.Errorf("failed to get namespace by name: %w", err)
		}
	}

	namespace, err := biz.NewNamespace(currentUserUid, req.Project.Name, req.Project.Desc)
	if err != nil {
		return nil, err
	}
	err = d.NamespaceUsecase.CreateNamespace(ctx, namespace, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("create  namespace failed: %w", err)
	}

	return &dmsV1.AddProjectReply{
		Payload: struct {
			//  namespace UID
			Uid string `json:"uid"`
		}{Uid: namespace.UID},
	}, nil
}

func (d *DMSService) DeleteNamespace(ctx context.Context, currentUserUid string, req *dmsV1.DelProjectReq) (err error) {
	err = d.NamespaceUsecase.DeleteNamespace(ctx, currentUserUid, req.ProjectUid)
	if err != nil {
		return fmt.Errorf("delete  namespace failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateNamespaceDesc(ctx context.Context, currentUserUid string, req *dmsV1.UpdateProjectReq) (err error) {
	err = d.NamespaceUsecase.UpdateNamespaceDesc(ctx, currentUserUid, req.ProjectUid, req.Project.Desc)
	if err != nil {
		return fmt.Errorf("update namespace failed: %w", err)
	}

	return nil
}

func (d *DMSService) ArchivedNamespace(ctx context.Context, currentUserUid string, req *dmsV1.ArchiveProjectReq) (err error) {
	err = d.NamespaceUsecase.ArchivedNamespace(ctx, currentUserUid, req.ProjectUid, true)
	if err != nil {
		return fmt.Errorf("archived namespace failed: %w", err)
	}

	return nil
}

func (d *DMSService) UnarchiveNamespace(ctx context.Context, currentUserUid string, req *dmsV1.UnarchiveProjectReq) (err error) {
	err = d.NamespaceUsecase.ArchivedNamespace(ctx, currentUserUid, req.ProjectUid, false)
	if err != nil {
		return fmt.Errorf("unarchive namespace failed: %w", err)
	}

	return nil
}
