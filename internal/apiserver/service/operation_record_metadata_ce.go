//go:build !enterprise

package service

import (
	"context"

	aV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func getOperationTypeNameListReply(context.Context) *aV1.GetOperationTypeNameListReply {
	return &aV1.GetOperationTypeNameListReply{Data: []aV1.OperationTypeNameListItem{}}
}

func getOperationActionListReply(context.Context) *aV1.GetOperationActionListReply {
	return &aV1.GetOperationActionListReply{Data: []aV1.OperationActionListItem{}}
}
