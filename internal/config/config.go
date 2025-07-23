package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type RedisConf struct {
	Addr         string `json:"addr"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

type Postgres struct {
	DSN          string `json:"dsn"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

type MachineryConf struct {
	RedisDB      int    `json:"redis_db"`
	DefaultQueue string `json:"default_queue"`
	ResultExpire int    `json:"result_expire"`
	Concurrency  int    `json:"concurrency"`
}

type Config struct {
	Postgres  Postgres      `json:"postgres"`
	Redis     RedisConf     `json:"redis"`
	Machinery MachineryConf `json:"machinery"`
}

var (
	cfg  Config
	once sync.Once
)

// Load 只会执行一次；env > JSON > 默认值
func Load() Config {
	once.Do(func() {
		// 直接读 JSON
		path := os.Getenv("CONFIG_PATH")
		if path == "" {
			path = "config.json" // 项目根目录
		}
		// 解析绝对路径
		abs, err := filepath.Abs(path)
		if err != nil {
			log.Panicf("FUCK 解析配置文件路径失败: %v", err)
		}
		// 打开文件
		file, err := os.Open(abs)
		if err != nil {
			log.Panicf("FUCK 打开 config.json 失败: %v", err)
		}
		defer file.Close()
		// 解析 JSON
		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			log.Panicf("FUCK 解析 config.json 失败: %v", err)
		}
		// 检查必需配置项
		if cfg.Postgres.DSN == "" {
			log.Panic("FUCK 配置项 postgres.dsn 为空")
		}
		if cfg.Redis.Addr == "" {
			log.Panic("FUCK 配置项 redis.addr 为空")
		}
	})

	return cfg
}
