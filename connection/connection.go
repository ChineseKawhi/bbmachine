package connection

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	lock      sync.Mutex
	Conn      *websocket.Conn
	WritingDb bool
}

func (c *Connection) WriteMessage(messageType int, data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

type ConnectionManager struct {
	lock        sync.RWMutex
	connections map[int64]*Connection
}

func NewConnectionManager() ConnectionManager {
	return ConnectionManager{connections: make(map[int64]*Connection)}
}

func (cm *ConnectionManager) Get(key int64) *Connection {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	fmt.Printf("Conn Get: %v, %v\n", key, cm.connections)
	if conn, ok := cm.connections[key]; ok {
		return conn
	}
	return nil
}

func (cm *ConnectionManager) Set(key int64, conn *Connection) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.connections[key] = conn
	fmt.Printf("Conn Set: %v, %v\n", key, cm.connections)
}

func (cm *ConnectionManager) Del(key int64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.connections, key)
	fmt.Printf("Conn Del: %v, %v\n", key, cm.connections)
}
