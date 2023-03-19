package auth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const (
	connectionPassword = "password123"
)

type AuthService struct {
	sync.RWMutex
	tokenCache map[string]bool
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func randToken() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func NewService() *AuthService {
	return &AuthService{
		tokenCache: make(map[string]bool),
	}
}

func (instance *AuthService) HandleAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupCorsResponse(w, r)
		if r.Method == http.MethodOptions {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&ErrorResponse{Code: http.StatusNotFound, Message: "not found"})
			return
		}
		password := r.Header.Get("X-Connection-Password")
		if password != connectionPassword {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&ErrorResponse{Code: http.StatusUnauthorized, Message: "invalid password"})
			return
		}
		token := randToken()
		instance.Lock()
		instance.tokenCache[token] = true
		instance.Unlock()
		data := AuthResponse{Token: token}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}

func (instance *AuthService) TokenIsValid(token string) bool {
	instance.RLock()
	defer instance.RUnlock()
	return instance.tokenCache[token]
}

func setupCorsResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, X-Connection-Password")
}
