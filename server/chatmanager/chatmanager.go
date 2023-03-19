package chatmanager

import (
	"beehive-chat/auth"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	incomingChatMessageType     = "chat:send"
	outgoingChatMessageType     = "chat:broadcast"
	outgoingChatListMessageType = "chat:list"
	broadcastMaxRetries         = 3
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
	// map of currently connected clients
	clients map[*websocket.Conn]bool
	// channel to add a client to clients
	add chan *websocket.Conn
	// channel to remove a client from clients
	remove chan *websocket.Conn
	// channel to broadcast a message to all clients
	broadcast chan Chat
	// map of all messages
	messages map[string]Chat
	// service to handle authorizing connections
	authService auth.AuthServicer
	// upgrader to handle upgrading socket connections
	upgrader *websocket.Upgrader
}

func NewManager(authService auth.AuthServicer, upgrader *websocket.Upgrader) *Manager {
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
	defaultCloseHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(code int, text string) error {
		log.Println("Closing connection to ", conn.RemoteAddr().String(), ": code ", code)
		err := defaultCloseHandler(code, text)
		if err != nil {
			log.Println("Error in default close handler: ", err.Error())
		}
		instance.remove <- conn
		return err
	})
	instance.add <- conn
	go instance.listenForMessages(conn)
}

func (instance *Manager) RemoveClient(conn *websocket.Conn) {
	instance.remove <- conn
}

func (instance *Manager) listenForMessages(conn *websocket.Conn) {
	for {
		var msg ChatEvent
		err := conn.ReadJSON(&msg)
		if err != nil {
			// if the normal close code was received, this error is fine and we know the
			// client closed the connection
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error reading json:", err.Error())
			}
			return
		}
		log.Println("Received message from", conn.RemoteAddr().String())
		log.Println(msg)
		if msg.MessageType == incomingChatMessageType {
			instance.broadcast <- msg.Data
		}
	}
}

func (instance *Manager) sendChatList(conn *websocket.Conn) error {
	chatListEvent := ChatListEvent{
		MessageType: outgoingChatListMessageType,
		Data:        instance.messages,
	}
	err := conn.WriteJSON(chatListEvent)
	if err != nil {
		// TODO: send error message to client
		log.Println("Error writing chat list to client: ", err.Error())
	}
	return err
}

func (instance *Manager) broadcastToClient(conn *websocket.Conn, chat ChatEvent) error {
	var err error
	for attempt := 0; attempt < broadcastMaxRetries; attempt++ {
		if attempt == 0 {
			err = conn.WriteJSON(chat)
		} else {
			err = instance.sendChatList(conn)
		}
		if err == nil {
			return nil
		}
		log.Printf("Attempt %d of %d - error writing message to client %s: %s", attempt+1, broadcastMaxRetries, conn.RemoteAddr().String(), err)
		time.Sleep(time.Second * (time.Duration(attempt) + 1))
	}
	return err
}

func (instance *Manager) Run() {
	for {
		select {
		case conn := <-instance.add:
			log.Println("Adding client:", conn.RemoteAddr().String())
			instance.clients[conn] = true
			instance.sendChatList(conn)
		case conn := <-instance.remove:
			log.Println("Removing client:", conn.RemoteAddr().String())
			if _, ok := instance.clients[conn]; ok {
				delete(instance.clients, conn)
				conn.Close()
			} else {
				log.Println("Error removing client", conn.RemoteAddr().String(), ": client not found")
			}
		case message := <-instance.broadcast:
			log.Println("Broadcasting: ", message)
			instance.messages[message.ChatId] = message
			chatEvent := ChatEvent{MessageType: outgoingChatMessageType, Data: message}
			for conn := range instance.clients {
				go instance.broadcastToClient(conn, chatEvent)
			}
		}
	}
}
