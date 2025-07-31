package websocket

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ConnectionPool 管理所有 WebSocket 客户端连接
// 通过 register/unregister/broadcast 通道处理连接的增删和消息分发

type ConnectionPool struct {
	clients    map[string]*Client            // 连接ID -> Client
	groups     map[string]map[string]*Client // 组名 -> (连接ID -> Client)
	Register   chan *Client                  // 新连接注册通道
	Unregister chan *Client                  // 连接注销通道
	broadcast  chan BroadcastMessage         // 广播任务通道
	mutex      sync.RWMutex                  // 保护 clients 和 groups
}

// NewConnectionPool 初始化连接池并启动管理协程
func NewConnectionPool() *ConnectionPool {
	p := &ConnectionPool{
		clients:    make(map[string]*Client),
		groups:     make(map[string]map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage),
	}
	// 启动后台事件循环
	go p.run()
	return p
}

// run 在单个 goroutine 中串行处理注册、注销、广播，减少锁竞争
func (p *ConnectionPool) run() {
	for {
		select {
		case client := <-p.Register:
			// 注册新客户端
			p.mutex.Lock()
			p.clients[client.ID] = client
			if client.Group != "" {
				if p.groups[client.Group] == nil {
					p.groups[client.Group] = make(map[string]*Client)
				}
				p.groups[client.Group][client.ID] = client
			}
			p.mutex.Unlock()
			log.Printf("新的 WebSocket 连接已注册: ID=%s, 用户=%s, 分组=%s\n", client.ID, client.UserID, client.Group)

		case client := <-p.Unregister:
			// 注销客户端，关闭发送通道
			p.mutex.Lock()
			_, existed := p.clients[client.ID]
			delete(p.clients, client.ID)
			if client.Group != "" {
				delete(p.groups[client.Group], client.ID)
			}
			p.mutex.Unlock()
			if existed {
				close(client.SendChan) // 关闭发送通道，通知发送协程退出
			}
			log.Printf("WebSocket 连接已注销: ID=%s, 用户=%s, 分组=%s\n", client.ID, client.UserID, client.Group)

		case msg := <-p.broadcast:
			// 分发广播消息
			if msg.Group != "" {
				// 分组广播
				p.mutex.RLock()
				groupClients := p.groups[msg.Group]
				p.mutex.RUnlock()

				for id, client := range groupClients {
					if id == msg.ExcludeID {
						continue
					}
					// 异步发送，缓冲满则移除慢客户端
					select {
					case client.SendChan <- msg.Message:
					default:
						log.Printf("WebSocket 连接发送缓冲满，移除客户端: ID=%s, 用户=%s, 分组=%s\n", client.ID, client.UserID, client.Group)
						p.removeClient(client)
					}
				}
			} else {
				// 全局广播
				p.mutex.RLock()
				allClients := make([]*Client, 0, len(p.clients))
				for _, c := range p.clients {
					allClients = append(allClients, c)
				}
				p.mutex.RUnlock()

				for _, client := range allClients {
					if client.ID == msg.ExcludeID {
						continue
					}
					select {
					case client.SendChan <- msg.Message:
					default:
						log.Printf("WebSocket 连接发送缓冲满，移除客户端: ID=%s, 用户=%s, 分组=%s\n", client.ID, client.UserID, client.Group)
						p.removeClient(client)
					}
				}
			}
		}
	}
}

// removeClient 从 pool 删除给定 client，不关闭其通道
func (p *ConnectionPool) removeClient(client *Client) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, existed := p.clients[client.ID]
	delete(p.clients, client.ID)
	if client.Group != "" {
		delete(p.groups[client.Group], client.ID)
	}
	if existed {
		close(client.SendChan) // 关闭发送通道，通知发送协程退出
	}
}

// BroadcastAll 全局广播消息
func (p *ConnectionPool) BroadcastAll(message []byte, excludeID string) {
	p.broadcast <- BroadcastMessage{Group: "", ExcludeID: excludeID, Message: message}
}

// BroadcastGroup 分组广播消息
func (p *ConnectionPool) BroadcastGroup(group string, message []byte, excludeID string) {
	p.broadcast <- BroadcastMessage{Group: group, ExcludeID: excludeID, Message: message}
}

// SendToID 发送消息到指定连接ID
func (p *ConnectionPool) SendToID(targetID string, message []byte) error {
	p.mutex.RLock()
	client, ok := p.clients[targetID]
	p.mutex.RUnlock()
	if !ok {
		return fiber.ErrNotFound
	}
	// 尝试异步写入，失败则移除
	select {
	case client.SendChan <- message:
		return nil
	default:
		p.removeClient(client)
		return fiber.ErrGone
	}
}

// GenerateID 使用 UUID 生成全局唯一连接ID
func (p *ConnectionPool) GenerateID() string {
	return uuid.New().String()
}
