package main

import (
	"log"
	"nextChatServer/internal/db"
	"os"
)

func init() {
	// 初始化日志文件
	logFileServer, err = os.OpenFile("log_server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}

	// 设置日志输出到文件
	log.SetOutput(logFileServer)

	// 记录日志文件初始化成功的消息
	log.Println("日志文件初始化成功")
}

func initDB() {
	db.MustGetDB()
	if err := db.AutoMigrate(); err != nil {
		panic(err)
	}
}

func shutdownDB() {
    sqlDB, _ := db.GetDB().DB()
    _ = sqlDB.Close()
}

func main() {
	// 程序主逻辑
	initDB()

	// 程序关闭逻辑
	defer shutdownDB()
}
