//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotBasicConfig = errors.New("personalization are enterprise version functions")

func (d *BasicUsecase) GetStaticLogo(ctx context.Context) (*BasicConfigParams, string, error) {
	return nil, "", errNotBasicConfig
}

func (d *BasicUsecase) Personalization(ctx context.Context, params *BasicConfigParams) error {
	return errNotBasicConfig
}
