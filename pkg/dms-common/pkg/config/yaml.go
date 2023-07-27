package config

import (
	"fmt"

	utilIo "github.com/actiontech/dms/pkg/dms-common/pkg/io"
	"github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"github.com/goccy/go-yaml"
)

func ParseYamlFile(logger log.Logger, filePath string, parseTo interface{}) error {
	bs, err := utilIo.ReadFile(logger, filePath)
	if nil != err {
		return fmt.Errorf("failed to read file: %s, error: %v", filePath, err)
	}
	err = yaml.Unmarshal(bs, parseTo)
	if nil != err {
		return fmt.Errorf("failed to parse yaml file: %s, error: %v", filePath, err)
	}

	err = Validate(parseTo)
	if nil != err {
		return fmt.Errorf("failed to validate yaml file: %s, error: %v", filePath, err)
	}
	return nil
}
