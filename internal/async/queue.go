package async

import (
	"fmt"
	"log"
	"sync"

	"github.com/RichardKnop/machinery/v1"
	machCfg "github.com/RichardKnop/machinery/v1/config"

	"nextChatServer/internal/config"
)

var (
	once   sync.Once
	server *machinery.Server
)

// GetServer 返回全局 *machinery.Server
func GetServer() *machinery.Server {
	once.Do(func() { initServer() })
	return server
}

// MustGetServer 初始化失败即 panic
func MustGetServer() *machinery.Server {
	s := GetServer()
	if s == nil {
		log.Fatalln("FUCK machinery server 未初始化")
	}
	return s
}

func initServer() {
	conf := config.Load()
	redisCfg := conf.Redis
	mach := conf.Machinery

	// Redis URL：redis://[:password]@host:port/db
	redisURL := fmt.Sprintf("redis://:%s@%s/%d",
		redisCfg.Password,
		redisCfg.Addr,
		mach.RedisDB,
	)

	c := &machCfg.Config{
		Broker:        redisURL,
		DefaultQueue:  mach.DefaultQueue,
		ResultBackend: redisURL,
		// 任务结果过期时间（秒）；0 表示永久保存
		ResultsExpireIn: mach.ResultExpire,
	}

	var err error
	server, err = machinery.NewServer(c)
	if err != nil {
		log.Fatalf("FUCK machinery.NewServer 错误: %v", err)
	}

	// 统一注册任务
	if err := server.RegisterTasks(map[string]interface{}{
		"add": Add, // 示例：加法
	}); err != nil {
		log.Fatalf("FUCK machinery 注册任务错误: %v", err)
	}
	log.Println("machinery 服务初始化完成")
}

/*------------------------------*
|   6.  Worker 启动封装        |
*------------------------------*/

// StartWorker 在当前进程启动一个 Worker；常在 main 或独立可执行文件里调用
func StartWorker(name string) {
	conf := config.Load().Machinery
	concurrency := conf.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}
	worker := GetServer().NewWorker(name, concurrency)
	if err := worker.Launch(); err != nil {
		log.Fatalf("FUCK machinery worker 启动错误: %v", err)
	}
}
