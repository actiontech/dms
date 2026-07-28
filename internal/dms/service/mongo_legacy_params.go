package service

import "strings"

// Deprecated Mongo additional_params removed from plugin meta; ignore if still present on old datasources/requests.
var deprecatedMongoAdditionalParams = map[string]struct{}{
	"auth_mechanism":    {},
	"tls":               {},
	"tls_skip_verify":   {},
	"direct_connection": {},
}

func isDeprecatedMongoAdditionalParam(dbType, name string) bool {
	if !strings.EqualFold(dbType, "MongoDB") {
		return false
	}
	_, ok := deprecatedMongoAdditionalParams[name]
	return ok
}
