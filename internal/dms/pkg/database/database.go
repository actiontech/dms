package database

import (
	"context"
)

type ConnectorImpl interface {
	IsConnectable(ctx context.Context) (bool, error)
}
