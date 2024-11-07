//go:build enterprise

package conf

func getOptimizationEnabled(opt *Options) bool {
	return opt.SQLE.OptimizationConfig.OptimizationKey != "" &&
		opt.SQLE.OptimizationConfig.OptimizationUrl != ""
}
