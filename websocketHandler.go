package main

import (
	"strings"
	"time"

	"log"
	ws "nextChatServer/internal/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// wsUpgradeMiddleware 校验并升级 WebSocket 请求
// 解析 query 参数 user 和 group，存入 c.Locals
func wsUpgradeMiddleware(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("userID", c.Query("user"))
		c.Locals("group", c.Query("group"))
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// websocketHandler 处理已升级的 WebSocket 连接
func websocketHandler(conn *websocket.Conn) {
	// 从上下文获取用户信息
	userID := conn.Locals("userID").(string)
	group := conn.Locals("group").(string)

	// 创建 Client 实例
	client := &ws.Client{
		ID:       pool.GenerateID(),
		UserID:   userID,
		Group:    group,
		Conn:     conn,
		SendChan: make(chan []byte, 256),
	}

	// 注册客户端
	pool.Register <- client
	// 断开时注销
	defer func() {
		pool.Unregister <- client
		_ = conn.Close()
	}()

	// 启动发送协程
	go func() {
		for msg := range client.SendChan {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("FUCK 发送消息失败: %v", err)
				return
			}
		}
	}()

	// 设置心跳：Pong 处理
	conn.SetPongHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	// 初始读超时
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 心跳 Ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("PING 发送失败，注销连接: userID=%s, group=%s, err=%v", userID, group, err)
				pool.Unregister <- client
				return
			}
		}
	}()

	// 读取循环
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break // 出错或关闭，退出循环
		}
		if mt != websocket.TextMessage {
			continue // 仅处理文本消息
		}
		text := string(msg)

		// 简单协议：prefix:payload
		parts := strings.SplitN(text, ":", 3)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "broadcast":
			// 全局广播，payload 即消息体
			pool.BroadcastAll([]byte(parts[1]), "")

		case "group":
			// 分组广播，parts[1]=组名，parts[2]=内容
			if len(parts) == 3 {
				pool.BroadcastGroup(parts[1], []byte(parts[2]), "")
			}

		case "to":
			// 定点发送，parts[1]=目标ID，parts[2]=内容
			if len(parts) == 3 {
				_ = pool.SendToID(parts[1], []byte(parts[2]))
			}

		default:
			// 默认：群组内广播，不回显自己
			if client.Group != "" {
				pool.BroadcastGroup(client.Group, msg, client.ID)
			} else {
				pool.BroadcastAll(msg, client.ID)
			}
		}
	}
}
