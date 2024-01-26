package biz

import (
	"sync"
	"time"
)

type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UIdWithName struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

// 数据源、用户缓存
var uidWithNameCacheCache UidWithNameCacheCache

type UidWithNameCacheCache struct {
	ulock     sync.Mutex
	UserCache map[string] /*uid*/ UIdWithName
	DBCache   map[string] /*uid*/ UIdWithName
}
