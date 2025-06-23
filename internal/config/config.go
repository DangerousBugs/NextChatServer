package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"log"
)

type Config struct {
	PostgresDSN string `json:"postgres_dsn"`
}

var (
	cfg  Config
	once sync.Once
)

// Load 只会执行一次；env > JSON > 默认值
func Load() Config {
	once.Do(func() {
		// 1. 先看环境变量
		if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
			cfg.PostgresDSN = dsn
			return
		}

		// 2. 再读 JSON
		path := os.Getenv("")
		if path == "" {
			path = "config.json" // 项目根目录
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			log.Panicf("解析配置文件路径失败: %v", err)
		}

		file, err := os.Open(abs)
		if err != nil {
			log.Panicf("打开 config.json 失败: %v", err)
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			log.Panicf("解析 config.json 失败: %v", err)
		}

		if cfg.PostgresDSN == "" {
			log.Panic("配置项 postgres_dsn 为空")
		}
	})

	return cfg
}
