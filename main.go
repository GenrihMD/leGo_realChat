package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fmt.Println("We begin...")

	http.HandleFunc("/", wsHandler)
	http.ListenAndServe(":8017", nil)
}
