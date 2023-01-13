package main

import "encoding/json"

type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan []byte
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

type Message struct {
	Sender   string `json:"sender,omitempty"`
	Receiver string `json:"receiver,omitempty"`
	Content  string `json:"content,omitempty"`
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			message, _ := json.Marshal(&Message{Content: "New user connected!" + client.userID})
			h.Send(message, client.userID)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
				message, _ := json.Marshal(&Message{Content: "User disconnected!" + client.userID})
				h.Send(message, client.userID)
			}
		case message := <-h.broadcast:
			m := &Message{}
			err := json.Unmarshal(message, &m)
			if err != nil {
				return
			}
			h.Send(message, m.Sender)

		}
	}
}

func (h *Hub) Send(message []byte, skip string) {
	for client := range h.clients {
		if client.userID != skip {
			client.send <- message
		}
	}
}
