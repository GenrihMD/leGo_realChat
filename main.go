package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}

		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		manager.broadcast <- jsonMessage
	}
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func (m *ClientManager) start() {
	for {
		select {
		case conn := <-m.register:
			m.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			m.send(jsonMessage, conn)

		case conn := <-m.unregister:
			if _, ok := m.clients[conn]; ok {
				close(conn.send)
				delete(m.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected"})
				m.send(jsonMessage, conn)
			}
		case message := <-m.broadcast:
			for conn := range m.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(m.clients, conn)
				}
			}
		}
	}
}

func (m *ClientManager) send(message []byte, ignore *Client) {
	for conn := range m.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
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
	u := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := u.Upgrade(w, r, nil)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	client := &Client{
		id:     uuid.NewV4().String(),
		socket: conn,
		send:   make(chan []byte),
	}

	manager.register <- client

	go client.read()
	go client.write()
}

func main() {
	fmt.Println("We begin...")

	http.HandleFunc("/", wsHandler)
	http.ListenAndServe(":8017", nil)
}
