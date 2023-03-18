package main

import (
	chatmanager "beehive-chat/chatmanager"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	connectionPassword = "password123"
)

var upgrader = websocket.Upgrader{}

func checkPassword(next http.Handler) http.Handler {
	// TODO: not plain text password
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := r.Header["X-Connection-Password"]
		log.Println(password)
		if len(password) != 1 || password[0] != connectionPassword {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	manager := chatmanager.NewManager()
	go manager.Run()
	mux := http.NewServeMux()
	connectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: make sure password is authenticate client
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading connection: ", err)
			return
		}
		manager.AddClient(conn)
	})
	mux.Handle("/", checkPassword(connectHandler))

	http.ListenAndServe("localhost:8081", mux)
}
