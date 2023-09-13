//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	pkgRand "github.com/actiontech/dms/pkg/rand"
)

func (d *BasicUsecase) GetStaticLogo(ctx context.Context) (*BasicConfigParams, string, error) {
	basicConfig, err := d.basicConfigRepo.GetBasicConfig(ctx)
	if err != nil {
		return nil, "", err
	}

	return basicConfig, http.DetectContentType(basicConfig.Logo), nil
}

const (
	MaxLogoSize = 1024 * 100 // 100KB
)

func (d *BasicUsecase) Personalization(ctx context.Context, params *BasicConfigParams) error {
	if params.File.Size > MaxLogoSize {
		return fmt.Errorf("image size exceeds %dKB", MaxLogoSize)
	}

	file, err := params.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buff, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if !d.IsImage(http.DetectContentType(buff)) {
		return errors.New("supports only (jpg, jpeg, png, gif) images")
	}

	params.Logo = buff

	basicConfig, err := d.basicConfigRepo.GetBasicConfig(ctx)
	if err != nil {
		return err
	}

	if basicConfig.UID == "" {
		uid, err := pkgRand.GenStrUid()
		if err != nil {
			return err
		}

		params.UID = uid
	} else {
		params.UID = basicConfig.UID
		params.CreatedAt = basicConfig.CreatedAt
	}

	return d.basicConfigRepo.SaveBasicConfig(ctx, params)
}

func (d *BasicUsecase) IsImage(fileType string) bool {
	switch fileType {
	case "image/jpeg", "image/jpg", "image/gif", "image/png":
		return true
	default:
		return false
	}
}
