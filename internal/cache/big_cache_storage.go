package cache

import (
	"context"
	"fmt"
	bigcache "github.com/allegro/bigcache/v3"
	"log"
	"time"
)

var globalCache *bigcache.BigCache

func init() {
	config := bigcache.Config{
		// 设置分区的数量，必须是2的整倍数
		Shards: 1024,
		// LifeWindow后,缓存对象被认为不活跃,但并不会删除对象
		LifeWindow: 5 * time.Minute,
		// CleanWindow后，会删除被认为不活跃的对象，0代表不操作；
		CleanWindow: 10 * time.Second,
		// 设置最大存储对象数量，仅在初始化时可以设置
		//MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntriesInWindow: 1,
		// 缓存对象的最大字节数，仅在初始化时可以设置
		MaxEntrySize: 500,
		// 是否打印内存分配信息
		Verbose: true,
		// 设置缓存最大值(单位为MB),0表示无限制
		HardMaxCacheSize: 8192,
		// 在缓存过期或者被删除时,可设置回调函数，参数是(key、val,reason)默认是nil不设置
		OnRemoveWithReason: nil,
	}
	var err error
	globalCache, err = bigcache.New(context.TODO(), config)
	globalCache.Close()
	if err != nil {
		log.Fatal(fmt.Printf("缓存初始化失败: %v", err))
	}
}

// Set 设置缓存项
func Set(key string, value []byte) error {
	return globalCache.Set(key, value)
}

// Get 获取缓存项
func Get(key string) ([]byte, error) {
	return globalCache.Get(key)
}

// Delete 删除缓存项
func Delete(key string) error {
	return globalCache.Delete(key)
}

// Close 关闭缓存
func Close() error {
	return globalCache.Close()
}