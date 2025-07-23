package main

import (
	"log"
	"nextChatServer/internal/async"
	"nextChatServer/internal/cache"
	"nextChatServer/internal/db"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
)

func init() {
	// 初始化日志文件
	logFileServer, err = os.OpenFile("log_server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("FUCK 打开文件失败: %v", err)
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

func initCache() {
	cache.MustGetRedis()
}

func initAsync() {
	// 初始化异步任务队列
	async.MustGetServer()

	// 开启worker进程
	go async.StartWorker("worker-local")
}

func shutdown() {
	sqlDB, _ := db.GetDB().DB()
	_ = sqlDB.Close()
	_ = cache.Close()
}

func main() {
	// 程序主逻辑
	initDB()
	initCache()
	initAsync()

	// 异步任务调用测试
	sig := &tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{Type: "int64", Value: 661},
			{Type: "int64", Value: 5},
		},
	}
	_, err := async.MustGetServer().SendTask(sig)
	if err != nil {
		log.Fatalf("发送异步任务失败: %v", err)
	} else {
		log.Println("异步任务发送成功")
		time.Sleep(20 * time.Second) // 等待任务执行完成
	}

	// 程序关闭逻辑
	defer shutdown()
}
