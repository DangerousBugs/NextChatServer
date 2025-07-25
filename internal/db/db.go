package db

import (
	"log"
	"os"
	"sync"
	"time"

	"nextChatServer/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	once sync.Once
	conn *gorm.DB
)

// GetDB 返回全局 *gorm.DB。首次调用时自动初始化。
func GetDB() *gorm.DB {
	once.Do(func() {
		initDB()
	})
	return conn
}

// MustGetDB 类似 GetDB，但初始化失败时直接 panic，适用于对数据库不可或缺的场景。
func MustGetDB() *gorm.DB {
	db := GetDB()
	if db == nil {
		panic("FUCK 数据库未初始化")
	}
	return db
}

// initDB 实际完成连接与连接池配置。
// 建议把 DSN 放在环境变量或配置中心里，这里简单读取。
func initDB() {
	// 读取配置
	// 通过 config 包获取
	dsn := config.Load().Postgres.DSN
	MaxIdleConns := config.Load().Postgres.MaxIdleConns
	MaxOpenConns := config.Load().Postgres.MaxOpenConns

	// 自定义日志级别（可根据需要替换）
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢查询阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("FUCK 连接数据库失败: %v", err)
	}

	// 连接池设置
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("FUCK 获取 gorm 的 sql.DB 失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(MaxIdleConns)        // 空闲连接数
	sqlDB.SetMaxOpenConns(MaxOpenConns)        // 打开连接数上限
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 单连接最大生命周期

	// 这里可以顺带做一次基础健康检查
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("FUCK 数据库 ping 错误: %v", err)
	}

	conn = db
	log.Println("数据库连接成功")
}
