package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"nextChatServer/internal/config"

	"github.com/redis/go-redis/v9" // go-redis v9
)

var (
	once   sync.Once
	client *redis.Client
)

// GetRedis 返回全局 *redis.Client
func GetRedis() *redis.Client {
	once.Do(func() { initRedis() })
	return client
}

// MustGetRedis 初始化失败即 panic
func MustGetRedis() *redis.Client {
	r := GetRedis()
	if r == nil {
		panic("Redis 未初始化")
	}
	return r
}

// initRedis 读取配置并建立连接
func initRedis() {
	conf := config.Load().Redis

	// 默认值兜底
	if conf.PoolSize == 0 {
		conf.PoolSize = 10
	}

	client = redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 简单连通性检查
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping 错误: %v", err)
	}
	log.Println("Redis 连接成功")
}

// Close 由 main/shutdown 调用
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}
