package chatmanager

import (
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Message  string `json:"message"`
	ClientId string `json:"clientId"`
}

type Manager struct {
	clients   map[*websocket.Conn]bool
	add       chan *websocket.Conn
	remove    chan *websocket.Conn
	broadcast chan *Message
	messages  []*Message
}

func NewManager() *Manager {
	return &Manager{
		clients:   make(map[*websocket.Conn]bool),
		add:       make(chan *websocket.Conn),
		remove:    make(chan *websocket.Conn),
		broadcast: make(chan *Message),
		messages:  []*Message{},
	}
}

func (instance *Manager) AddClient(conn *websocket.Conn) {
	instance.add <- conn
	go instance.listenForMessages(conn)
	// TODO: send list of messages
}

func (instance *Manager) RemoveClient(conn *websocket.Conn) {
	instance.remove <- conn
}

func (instance *Manager) listenForMessages(conn *websocket.Conn) {
	log.Println("Listening for messages:", conn.RemoteAddr().String())
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json:", err.Error())
			instance.remove <- conn
			return
		}
		log.Println("Received message")
		log.Println(msg)
		instance.broadcast <- &msg
	}
}

func (instance *Manager) Run() {
	for {
		select {
		case client := <-instance.add:
			log.Println("Adding client:", client.RemoteAddr().String())
			instance.clients[client] = true
		case client := <-instance.remove:
			log.Println("Removing client:", client.RemoteAddr().String())
			if _, ok := instance.clients[client]; ok {
				delete(instance.clients, client)
				client.Close()
			} else {
				log.Println("Error removing client", client.RemoteAddr().String(), ": client not found")
			}
		case message := <-instance.broadcast:
			log.Println("Broadcasting: ", message)
			instance.messages = append(instance.messages, message)
			for client := range instance.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Println("Error writing message to client", client.RemoteAddr().String(), ": ", err)
				}
			}
		}
	}
}
