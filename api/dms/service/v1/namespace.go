package v1

import (
	"fmt"

	base "github.com/actiontech/dms/api/base/v1"
)

// A Namespace
type Namespace struct {
	// namespace name
	Name string `json:"name"`
	// namespace desc
	Desc string `json:"desc"`
}

// swagger:parameters AddNamespace
type AddNamespaceReq struct {
	// Add new Namespace
	// in:body
	Namespace *Namespace `json:"namespace" validate:"required"`
}

func (u *AddNamespaceReq) String() string {
	if u == nil {
		return "AddNamespaceReq{nil}"
	}
	return fmt.Sprintf("AddNamespaceReq{NamespaceName:%s}", u.Namespace.Name)
}

// swagger:model AddNamespaceReply
type AddNamespaceReply struct {
	// Add Namespace reply
	Payload struct {
		// Namespace UID
		Uid string `json:"uid"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

func (u *AddNamespaceReply) String() string {
	if u == nil {
		return "AddNamespaceReply{nil}"
	}
	return fmt.Sprintf("AddNamespaceReply{Uid:%s}", u.Payload.Uid)
}

// swagger:parameters DelNamespace
type DelNamespaceReq struct {
	// namespace uid
	// in:path
	NamespaceUid string `param:"namespace_uid" json:"namespace_uid" validate:"required"`
}

func (u *DelNamespaceReq) String() string {
	if u == nil {
		return "DelNamespaceReq{nil}"
	}
	return fmt.Sprintf("DelNamespaceReq{Uid:%s}", u.NamespaceUid)
}

type UpdateNamespace struct {
	// Namespace desc
	Desc *string `json:"desc"`
}

// swagger:parameters UpdateNamespace
type UpdateNamespaceReq struct {
	// Namespace uid
	// Required: true
	// in:path
	NamespaceUid string `param:"namespace_uid" json:"namespace_uid" validate:"required"`
	// Update a namespace
	// in:body
	Namespace *UpdateNamespace `json:"namespace" validate:"required"`
}

func (u *UpdateNamespaceReq) String() string {
	if u == nil {
		return "UpdateNamespaceReq{nil}"
	}
	if u.Namespace == nil {
		return "UpdateNamespaceReq{Namespace:nil}"
	}
	return fmt.Sprintf("UpdateNamespaceReq{Uid:%s}", u.NamespaceUid)
}

// swagger:parameters ArchiveNamespace
type ArchiveNamespaceReq struct {
	// Namespace uid
	// Required: true
	// in:path
	NamespaceUid string `param:"namespace_uid" json:"namespace_uid" validate:"required"`
}

// swagger:parameters UnarchiveNamespace
type UnarchiveNamespaceReq struct {
	// Namespace uid
	// Required: true
	// in:path
	NamespaceUid string `param:"namespace_uid" json:"namespace_uid" validate:"required"`
}
