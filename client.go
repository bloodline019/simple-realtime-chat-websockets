package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type Client struct {
	userID string
	hub    *Hub
	conn   *websocket.Conn

	send chan []byte // Канал для отправки сообщений
}

var upgrader = websocket.Upgrader{
	// Позволяем открывать websocket соединение с chrome/extensions
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(hub *Hub, respWriter http.ResponseWriter, request *http.Request) {
	// Дополняем GET запрос до websocket соединения
	conn, err := upgrader.Upgrade(respWriter, request, nil)
	if err != nil {
		http.NotFound(respWriter, request)
		return
	}
	client := &Client{
		userID: uuid.NewV1().String(),
		conn:   conn,
		send:   make(chan []byte),
		hub:    hub,
	}

	hub.register <- client

	go client.read()
	go client.write()

}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			c.conn.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.userID, Content: string(message)})
		c.hub.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
