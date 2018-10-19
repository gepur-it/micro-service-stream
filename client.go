package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	subscribe Subscribe
}

type Subscribe struct {
	ApiKey   string `json:"apiKey"`
}


