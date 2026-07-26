package testutil

import (
	"fmt"
	"sync"

	"github.com/bfenetworks/bfe/bfe_util/redis_client"
)

// MockRedisClient 内存 Redis Mock 实现
// 实现 redis_client.Client 接口，用于测试环境替代真实 Redis
type MockRedisClient struct {
	mu   sync.Mutex
	data map[string]int64
}

// 确保 MockRedisClient 实现了 redis_client.Client 接口
var _ redis_client.Client = (*MockRedisClient)(nil)

// NewMockRedisClient 创建新的 Mock Redis 客户端
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]int64),
	}
}

// Setex 设置 key 的值并设置过期时间（mock 忽略过期时间）
func (m *MockRedisClient) Setex(key string, value []byte, expire int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = 0
	return nil
}

// Get 获取 key 的值
func (m *MockRedisClient) Get(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return nil, nil
}

// Expire 设置 key 的过期时间（mock 忽略）
func (m *MockRedisClient) Expire(key string, expire int) error {
	return nil
}

// Incr 将 key 对应的值增加 1
func (m *MockRedisClient) Incr(key string) (int64, error) {
	return m.IncrBy(key, 1)
}

// IncrAndExpire Incr 并设置过期时间（mock 忽略过期时间）
func (m *MockRedisClient) IncrAndExpire(key string, expire int) (int64, error) {
	return m.IncrBy(key, 1)
}

// Decr 将 key 对应的值减少 1
func (m *MockRedisClient) Decr(key string) (int64, error) {
	return m.IncrBy(key, -1)
}

// PIncr 批量增加多个 key
func (m *MockRedisClient) PIncr(keys []string) ([]int64, error) {
	results := make([]int64, len(keys))
	for i, key := range keys {
		val, err := m.IncrBy(key, 1)
		if err != nil {
			return nil, err
		}
		results[i] = val
	}
	return results, nil
}

// GetInt64 获取 key 对应的 int64 值
func (m *MockRedisClient) GetInt64(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key], nil
}

// IncrBy 将 key 对应的值增加 delta
func (m *MockRedisClient) IncrBy(key string, delta int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] += delta
	return m.data[key], nil
}

// NewScript 创建新的 RedisScript（mock 实现）
func (m *MockRedisClient) NewScript(src string) redis_client.RedisScript {
	return &MockRedisScript{src: src, client: m}
}

// Reset 清空所有数据
func (m *MockRedisClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]int64)
}

// MockRedisScript Redis Script mock 实现
type MockRedisScript struct {
	src    string
	client *MockRedisClient
}

// 确保 MockRedisScript 实现了 redis_client.RedisScript 接口
var _ redis_client.RedisScript = (*MockRedisScript)(nil)

// Run 执行脚本（mock 实现：模拟 Lua 配额检查脚本）
func (s *MockRedisScript) Run(key string, args ...interface{}) (interface{}, error) {
	if len(args) > 0 {
		quota, ok := args[0].(int64)
		if !ok {
			return nil, fmt.Errorf("mock script: invalid quota argument type")
		}
		balance, _ := s.client.GetInt64(key)
		if balance >= quota {
			return s.client.IncrBy(key, quota)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("mock script: no arguments provided")
}