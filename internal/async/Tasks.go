package async

import (
	"log"
	"time"
)

/*------------------------------*
|   4.  任务函数示例            |
*------------------------------*/

// Add (示例任务)：两数相加
func Add(a, b int64) (int64, error) {
	log.Printf("执行 Add 任务: %d + %d", a, b)
	// 模拟一些处理时间
	time.Sleep(2 * time.Second)
	log.Printf("Add 任务结果: %d", a+b)
	return a + b, nil
}
