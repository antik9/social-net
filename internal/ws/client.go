package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/pkg/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	user   *models.User
	friend *models.User
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		go saveMessage(string(message), c.user.Id, c.friend.Id)
		c.hub.broadcast <- formatMessage(string(message), c.user.FirstName)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func formatMessage(message, userName string) []byte {
	now := time.Now().Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf(
		"<div>%s</div><div>%s: %s</div><br/>",
		now, userName, message,
	))
}

func ServeWs(hubs *Hubs, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	key, reversedKey := getKeysFromRequest(r)
	hubs.Lock()
	hub, ok := hubs.connections[key]
	if !ok {
		hub = newHub()
		go hub.run()
		hubs.connections[key] = hub
		hubs.connections[reversedKey] = hub
	}
	hubs.Unlock()

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		user:   getUserFromRequest(r, "user1"),
		friend: getUserFromRequest(r, "user2"),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func getKeysFromRequest(r *http.Request) (string, string) {
	user1, user2 := r.URL.Query().Get("user1"), r.URL.Query().Get("user2")
	return user1 + "<->" + user2, user2 + "<->" + user1
}

func getUserFromRequest(r *http.Request, userKey string) *models.User {
	userId := r.URL.Query().Get(userKey)
	id, _ := strconv.Atoi(userId)
	return models.GetUserById(id)
}

func saveMessage(message string, userId, friendId int) {
	data, _ := json.Marshal(map[string]interface{}{
		"message":  message,
		"userId":   userId,
		"friendId": friendId,
	})

	http.Post(fmt.Sprintf(
		"http://%s:%s/chat/message",
		config.Conf.ChatServer.Host, config.Conf.ChatServer.Port,
	), "application/json", bytes.NewBuffer(data))
}
