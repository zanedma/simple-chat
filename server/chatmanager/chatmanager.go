package chatmanager

import (
	"beehive-chat/auth"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	incomingChatMessageType     = "chat:send"
	outgoingChatMessageType     = "chat:broadcast"
	outgoingChatListMessageType = "chat:list"
)

type Chat struct {
	Data      string `json:"data"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
	ChatId    string `json:"chatId"`
}

type ChatEvent struct {
	MessageType string `json:"messageType"`
	Data        Chat   `json:"data"`
}

type ChatListEvent struct {
	MessageType string          `json:"messageType"`
	Data        map[string]Chat `json:"data"`
}

type Manager struct {
	clients   map[*websocket.Conn]bool
	add       chan *websocket.Conn
	remove    chan *websocket.Conn
	broadcast chan Chat

	messages    map[string]Chat
	authService *auth.AuthService
	upgrader    *websocket.Upgrader
}

func NewManager(authService *auth.AuthService, upgrader *websocket.Upgrader) *Manager {
	return &Manager{
		clients:     make(map[*websocket.Conn]bool),
		add:         make(chan *websocket.Conn),
		remove:      make(chan *websocket.Conn),
		broadcast:   make(chan Chat),
		messages:    map[string]Chat{},
		authService: authService,
		upgrader:    upgrader,
	}
}

func (instance *Manager) HandleConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if !instance.authService.TokenIsValid(token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		conn, err := instance.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading connection:", err)
			return
		}
		instance.addClient(conn)
	}
}

func (instance *Manager) addClient(conn *websocket.Conn) {
	instance.add <- conn
	go instance.listenForMessages(conn)
	// TODO: send list of messages
}

func (instance *Manager) RemoveClient(conn *websocket.Conn) {
	instance.remove <- conn
}

func (instance *Manager) listenForMessages(conn *websocket.Conn) {
	for {
		var msg ChatEvent
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json:", err.Error())
			// TODO: send error response to client
			return
		}
		log.Println("Received message from", conn.RemoteAddr().String())
		log.Println(msg)
		if msg.MessageType == incomingChatMessageType {
			instance.broadcast <- msg.Data
		}
	}
}

func (instance *Manager) Run() {
	for {
		select {
		case client := <-instance.add:
			log.Println("Adding client:", client.RemoteAddr().String())
			instance.clients[client] = true
			chatListEvent := ChatListEvent{
				MessageType: outgoingChatListMessageType,
				Data:        instance.messages,
			}
			err := client.WriteJSON(chatListEvent)
			if err != nil {
				// TODO: send error message to client
				log.Println("Error writing chat list to client: ", err.Error())
				return
			}
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
			instance.messages[message.ChatId] = message
			chatEvent := ChatEvent{MessageType: outgoingChatMessageType, Data: message}
			for client := range instance.clients {
				err := client.WriteJSON(chatEvent)
				if err != nil {
					log.Println("Error writing message to client", client.RemoteAddr().String(), ": ", err)
				}
			}
		}
	}
}
