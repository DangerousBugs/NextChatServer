package websocket

import (
	"github.com/gofiber/contrib/websocket"
)

// Client 表示一个 WebSocket 客户端连接
// 包含连接 ID、用户 ID、所属分组、底层连接对象及发送通道
// 发送通道用于异步写，避免并发写入冲突

type Client struct {
	ID       string           // 连接唯一标识
	UserID   string           // 客户端用户标识
	Group    string           // 客户端所属分组，空串表示未加入任何组
	Conn     *websocket.Conn  // 原生 WebSocket 连接对象
	SendChan chan []byte      // 发送消息缓冲通道
}

// BroadcastMessage 定义广播任务
// Group: 目标分组，空串表示全局广播
// ExcludeID: 可选要排除的连接ID
// Message: 待发送的消息内容

type BroadcastMessage struct {
	Group     string
	ExcludeID string
	Message   []byte
}
