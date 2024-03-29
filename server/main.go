package main

import (
	"simple-chat/auth"
	chatmanager "simple-chat/chatmanager"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	connectionPassword = "password123"
	allowedOrigin      = "http://localhost:3000"
	port               = 8081
)

func checkPassword(next http.Handler) http.Handler {
	// TODO: not plain text password
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := r.Header["X-Connection-Password"]
		if len(password) != 1 || password[0] != connectionPassword {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	authService := auth.NewService()
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == allowedOrigin
		},
	}
	manager := chatmanager.NewManager(authService, upgrader)
	go manager.Run()
	http.Handle("/auth", authService.HandleAuth())
	http.Handle("/chat", manager.HandleConnection())

	log.Println("Starting server on port", port)
	err := http.ListenAndServe("localhost:8081", nil)
	if err != nil {
		log.Println("Error starting server: ", err.Error())
	}
}
