//go:build !enterprise

package service

import (
	"context"
	"errors"
)

var errNotSupportImportDBServices = errors.New("ImportDBServices related functions are enterprise version functions")

func (d *DMSService) importDBServices(ctx context.Context, userUid, projectUid, fileContent string) ([]byte, error) {
	return nil, errNotSupportProject
}
