package biz

import (
	"sync"
)

// FunctionSupportProvider 功能支持提供者接口
// 各功能模块实现此接口，向注册中心提供自己支持的数据库类型
type FunctionSupportProvider interface {
	// GetFunctionName 返回功能名称，如 "data_masking"
	GetFunctionName() string
	// GetSupportedDBTypes 返回支持的数据库类型列表
	GetSupportedDBTypes() []string
}

// FunctionSupportRegistry 功能支持注册中心
// 管理各功能模块支持的数据库类型，用于全局查询
type FunctionSupportRegistry struct {
	mu        sync.RWMutex
	providers map[string]FunctionSupportProvider
}

// NewFunctionSupportRegistry 创建功能支持注册中心
func NewFunctionSupportRegistry() *FunctionSupportRegistry {
	return &FunctionSupportRegistry{
		providers: make(map[string]FunctionSupportProvider),
	}
}

// Register 注册功能支持提供者
func (r *FunctionSupportRegistry) Register(provider FunctionSupportProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.GetFunctionName()] = provider
}

// GetSupportedDBTypes 获取指定功能支持的数据库类型列表
// 如果功能未注册，返回 nil
func (r *FunctionSupportRegistry) GetSupportedDBTypes(functionName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[functionName]
	if !ok {
		return nil
	}
	return provider.GetSupportedDBTypes()
}
