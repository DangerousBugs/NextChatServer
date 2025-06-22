package main

import (
	"os"
	"log"
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

func main() {
	// 程序主逻辑
	log.Println("程序开始运行")
}
