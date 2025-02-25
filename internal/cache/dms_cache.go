package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var DmsCache *cache.Cache
func init () {
	DmsCache = cache.New(cache.NoExpiration, 10 * time.Minute)
}
