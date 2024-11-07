//go:build !enterprise

package conf

func getOptimizationEnabled(opt *Options) bool {
	return false
}
