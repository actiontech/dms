package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type NamespaceStatus string

const (
	NamespaceStatusArchived NamespaceStatus = "archived"
	NamespaceStatusActive   NamespaceStatus = "active"
	NamespaceStatusUnknown  NamespaceStatus = "unknown"
)

type Namespace struct {
	Base

	UID           string
	Name          string
	Desc          string
	CreateUserUID string
	CreateTime    time.Time
	Status        NamespaceStatus
}

func NewNamespace(createUserUID, name, desc string) (*Namespace, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &Namespace{
		UID:           uid,
		Name:          name,
		Desc:          desc,
		Status:        NamespaceStatusActive,
		CreateUserUID: createUserUID,
	}, nil
}

func initNamespaces() []*Namespace {
	return []*Namespace{
		{
			UID:           pkgConst.UIDOfNamespaceDefault,
			Name:          "default",
			Desc:          "default namespace",
			Status:        NamespaceStatusActive,
			CreateUserUID: pkgConst.UIDOfUserAdmin,
		},
	}
}

type NamespaceRepo interface {
	SaveNamespace(ctx context.Context, namespace *Namespace) error
	ListNamespaces(ctx context.Context, opt *ListNamespacesOption, currentUserUID string) (namespaces []*Namespace, total int64, err error)
	GetNamespace(ctx context.Context, namespaceUid string) (*Namespace, error)
	GetNamespaceByName(ctx context.Context, namespaceName string) (*Namespace, error)
	UpdateNamespace(ctx context.Context, u *Namespace) error
	DelNamespace(ctx context.Context, namespaceUid string) error
}

type NamespaceUsecase struct {
	tx                        TransactionGenerator
	repo                      NamespaceRepo
	memberUsecase             *MemberUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	pluginUsecase             *PluginUsecase
	log                       *utilLog.Helper
}

func NewNamespaceUsecase(log utilLog.Logger, tx TransactionGenerator, repo NamespaceRepo, memberUsecase *MemberUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase, pluginUsecase *PluginUsecase) *NamespaceUsecase {
	return &NamespaceUsecase{
		tx:                        tx,
		repo:                      repo,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.namespace")),
		memberUsecase:             memberUsecase,
		pluginUsecase:             pluginUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
	}
}

type ListNamespacesOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      NamespaceField
	FilterBy     []pkgConst.FilterCondition
}

func (d *NamespaceUsecase) ListNamespace(ctx context.Context, option *ListNamespacesOption, currentUserUid string) (namespaces []*Namespace, total int64, err error) {
	// filter visible namespce space in advance
	// user can only view his belonging namespace,sys user can view all namespace
	if currentUserUid != pkgConst.UIDOfUserSys {
		namespaceWithOppermissions, err := d.opPermissionVerifyUsecase.GetUserNamespaceOpPermission(ctx, currentUserUid)
		if err != nil {
			return nil, 0, err
		}
		canViewableId := make([]string, 0)
		for _, namespaceWithOppermission := range namespaceWithOppermissions {
			canViewableId = append(canViewableId, namespaceWithOppermission.NamespaceUid)
		}
		option.FilterBy = append(option.FilterBy, pkgConst.FilterCondition{
			Field:    string(NamespaceFieldUID),
			Operator: pkgConst.FilterOperatorIn,
			Value:    canViewableId,
		})

	}

	namespaces, total, err = d.repo.ListNamespaces(ctx, option, currentUserUid)
	if err != nil {
		return nil, 0, fmt.Errorf("list namespaces failed: %v", err)
	}

	return namespaces, total, nil
}

func (d *NamespaceUsecase) InitNamespaces(ctx context.Context) (err error) {
	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()
	for _, n := range initNamespaces() {

		_, err := d.GetNamespace(ctx, n.UID)
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get namespace: %v", err)
		}

		// not exist, then create it.
		err = d.repo.SaveNamespace(tx, n)
		if err != nil {
			return fmt.Errorf("save namespaces failed: %v", err)
		}

		_, err = d.memberUsecase.AddUserToNamespaceAdminMember(tx, pkgConst.UIDOfUserAdmin, n.UID)
		if err != nil {
			return fmt.Errorf("add admin to namespaces failed: %v", err)
		}
	}
	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	d.log.Debug("init namespace success")
	return nil
}

func (d *NamespaceUsecase) GetNamespace(ctx context.Context, namespaceUid string) (*Namespace, error) {
	return d.repo.GetNamespace(ctx, namespaceUid)
}
