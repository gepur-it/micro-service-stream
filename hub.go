package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	subscribe  chan QueryCommand
	register   chan *Client
	unregister chan *Client
}

func hub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		subscribe:  make(chan QueryCommand),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		if len(client.subscribe.Id) != 0 {
			client.send <- message
		} else {
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) subscribeMessage(message QueryCommand) {
	for client := range h.clients {
		if len(client.subscribe.Id) != 0 {
			for recipient := range message.Recipients {
				if message.Recipients[recipient] == client.subscribe.UserID {
					bytesToSend, err := json.Marshal(message)

					if err != nil {
						logger.WithFields(logrus.Fields{
							"error": err,
						}).Error("Can`t encode subscribe message:")
					}

					client.send <- bytesToSend
				}
			}
		} else {
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case message := <-h.subscribe:
			h.subscribeMessage(message)
		}
	}
}
