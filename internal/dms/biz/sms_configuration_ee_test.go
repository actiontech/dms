//go:build enterprise

package biz

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestGenerateStatelessCode(t *testing.T) {


	key := "test_key"

	tests := []struct {
		name     string
		needEaual bool            // 模拟的时间戳
		key      string           // 测试密钥
		wantCode string           // 期望的验证码
	}{
		{
			name:     "当前时间段生成验证码",
			needEaual: true,
			key:      key,
			wantCode: calculateExpectedCode(key, time.Now().Unix()),
		},
		{
			name:     "不同时间段生成不同验证码",
			needEaual: false,
			key:      key,
			wantCode: calculateExpectedCode(key, time.Now().Unix()+300),
		},
		{
			name:     "不同密钥生成不同验证码",
			needEaual: false,
			key:      "different_key_1",
			wantCode: calculateExpectedCode("different_key_2", time.Now().Unix()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := GenerateStatelessCode(tt.key)
			if tt.needEaual {
				if got != tt.wantCode {
					t.Errorf("GenerateStatelessCode() = %v, want %v", got, tt.wantCode)
				}
			} else {
				if got == tt.wantCode {
					t.Errorf("GenerateStatelessCode() = %v, want %v", got, tt.wantCode)
				}
			}
		})
	}
}

// 辅助函数：计算预期验证码
func calculateExpectedCode(key string, timestamp int64) string {
	timeSlot := timestamp / VerifyCodeTimeSlot
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeSlot))

	h := hmac.New(sha256.New, []byte(key))
	h.Write(timeBytes)
	hash := h.Sum(nil)

	num := binary.BigEndian.Uint32(hash[len(hash)-4:])
	code := (num % 9000) + 1000
	return fmt.Sprintf("%04d", code)
}


func TestValidateStatelessCode(t *testing.T) {

	key := "test_key"
	currentTime := time.Now().Unix()

	// 生成当前和上一个时间段的验证码
	currentCode := GenerateStatelessCode(key)
	prevCode := calculateExpectedCode(key, currentTime-VerifyCodeTimeSlot)

	tests := []struct {
		name        string
		inputCode   string
		key         string
		wantIsValid bool
	}{
		{
			name:        "当前时间段验证码正确",
			inputCode:   currentCode,
			key:         key,
			wantIsValid: true,
		},
		{
			name:        "上一个时间段验证码正确",
			inputCode:   prevCode,
			key:         key,
			wantIsValid: true,
		},
		{
			name:        "错误验证码",
			inputCode:   "1234",
			key:         key,
			wantIsValid: false,
		},
		{
			name:        "不同密钥无效",
			inputCode:   currentCode,
			key:         "wrong_key",
			wantIsValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateStatelessCode(tt.key, tt.inputCode)
			if got != tt.wantIsValid {
				t.Errorf("ValidateStatelessCode() = %v, want %v", got, tt.wantIsValid)
			}
		})
	}
}

func TestCheckSmsRateLimit(t *testing.T) {
	// 使用独立缓存实例避免测试干扰
	testCache:=	cache.New(cache.NoExpiration, 10 * time.Minute)
	
	phone := "13800138000"
	key := "test_key"

	tests := []struct {
		name        string
		preAction   func()          // 测试前置操作
		key         string
		phone       string
		wantErr     error
	}{
		{
			name:        "首次请求通过",
			preAction:   func() { testCache.Flush() },
			key:         key,
			phone:       phone,
			wantErr:     nil,
		},
		{
			name:        "第二次请求被限流",
			preAction:   func() { testCache.Set(getCacheKey(key, phone), true, 1*time.Minute) },
			key:         key,
			phone:       phone,
			wantErr:     ErrTooFrequentMinute,
		},
		{
			name:        "不同手机号不限流",
			preAction:   func() { testCache.Set(getCacheKey(key, phone), true, 1*time.Minute) },
			key:         key,
			phone:       "other_phone",
			wantErr:     nil,
		},
		{
			name:        "缓存过期后请求通过",
			preAction:   func() {
				testCache.Set(getCacheKey(key, phone), true, 1*time.Millisecond)
				time.Sleep(2 * time.Millisecond) // 等待缓存过期
			},
			key:         key,
			phone:       phone,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置缓存状态
			tt.preAction()
			
			// 替换全局缓存为测试实例
			originalCache := verifyCodeCache
			verifyCodeCache = testCache
			defer func() { verifyCodeCache = originalCache }()

			err := checkSmsRateLimit(tt.key, tt.phone)
			if err != tt.wantErr {
				t.Errorf("checkSmsRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 辅助函数：生成缓存键
func getCacheKey(key, phone string) string {
	return fmt.Sprintf("%s:%s",  key, phone)
}