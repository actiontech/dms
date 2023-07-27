//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportNamespace = errors.New("namespace related functions are enterprise version functions")

func (d *NamespaceUsecase) CreateNamespace(ctx context.Context, namespace *Namespace, createUserUID string) (err error) {
	return errNotSupportNamespace
}

func (d *NamespaceUsecase) GetNamespace(ctx context.Context, namespaceUid string) (*Namespace, error) {
	return nil, errNotSupportNamespace
}

func (d *NamespaceUsecase) GetNamespaceByName(ctx context.Context, namespaceName string) (*Namespace, error) {
	return nil, errNotSupportNamespace
}

func (d *NamespaceUsecase) UpdateNamespaceDesc(ctx context.Context, currentUserUid, namespaceUid string, desc *string) (err error) {

	return errNotSupportNamespace
}

func (d *NamespaceUsecase) ArchivedNamespace(ctx context.Context, currentUserUid, namespaceUid string, archived bool) (err error) {

	return errNotSupportNamespace
}

func (d *NamespaceUsecase) DeleteNamespace(ctx context.Context, currentUserUid, namespaceUid string) (err error) {
	return errNotSupportNamespace
}

func (d *NamespaceUsecase) isNamespaceActive(ctx context.Context, namespaceUid string) error {
	return nil
}
