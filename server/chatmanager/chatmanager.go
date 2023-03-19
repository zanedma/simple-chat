package chatmanager

import (
	"beehive-chat/auth"
	"fmt"
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

// handle a connection request from a client. First checks the token to authenticate the request
// and if valid, upgrades the connection and adds the client to the chat manager
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
		instance.addClient(conn, token)
	}
}

// add a connected client to the chat manager.
func (instance *Manager) addClient(conn *websocket.Conn, token string) {
	defaultCloseHandler := conn.CloseHandler()

	// set custom behavior when a connection closes
	conn.SetCloseHandler(func(code int, text string) error {
		log.Println("Closing connection to ", conn.RemoteAddr().String(), ": code ", code)
		err := defaultCloseHandler(code, text)
		if err != nil {
			log.Println("Error in default close handler: ", err.Error())
		}
		instance.remove <- conn
		// invalidate the token when the connection closes. If we want to add automatic reconnect
		// functionality to to the frontend on unexpected close, this would need to be changed to
		// check the close code or something else
		instance.authService.RemoveToken(token)
		return err
	})

	instance.add <- conn
	go instance.listenForMessages(conn)
}

// read incoming messages and handle them accordingly (currently only 1 message type is incoming
// client side)
func (instance *Manager) listenForMessages(conn *websocket.Conn) {
	for {
		var msg ChatEvent
		err := conn.ReadJSON(&msg)
		if err != nil {
			// if the normal close code was received, this error is fine and we know the
			// client closed the connection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error reading json:", err.Error())
			}
			return
		}
		log.Println("Received message from", conn.RemoteAddr().String())
		if msg.MessageType == incomingChatMessageType {
			instance.broadcast <- msg.Data
		}
	}
}

// send the full list of chats to the client
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

// broadcast a message to a client. If an error occurrs, the manager will retry after waiting
// a backed off amount of time. On attempts after the first, the full list will be sent instead
// of the single message to make sure the client has the most up-to-date data. If the retry limit
// is hit, the connection will be closed, and the client will have to reconnect. On reconnect,
// it will receive the most up-to-date chat list.
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
	conn.CloseHandler()(websocket.CloseInternalServerErr, "broadcast error")
	return fmt.Errorf("hello")
}

// run the chat manager and listen on the channels for events
func (instance *Manager) Run() {
	for {
		select {
		case conn := <-instance.add:
			log.Println("Adding client:", conn.RemoteAddr().String())
			instance.clients[conn] = true
			err := instance.sendChatList(conn)
			if err != nil {
				conn.CloseHandler()(websocket.CloseInternalServerErr, "error sending chat list")
			}
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
