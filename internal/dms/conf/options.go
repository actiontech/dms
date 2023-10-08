package conf

type Options struct {
	Database struct {
		UserName string `yaml:"username" `
		Password string `yaml:"password" `
		Host     string `yaml:"host" validate:"required"`
		Port     string `yaml:"port" validate:"required"`
		Database string `yaml:"database" validate:"required"`
		Debug    bool   `yaml:"debug"`
	} `yaml:"database"`
}

// func ReadOptions(log utilLog.Logger, path string) (*Options, error) {
// 	var opts Options
// 	if err := utilConf.ParseYamlFile(log, path, &opts); err != nil {
// 		return nil, err
// 	}
// 	return &opts, nil
// }
