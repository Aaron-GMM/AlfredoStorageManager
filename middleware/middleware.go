package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"alfredostoragemanager/api/models"
)

var (
	authUsername string
	authPassword string

	sessions = make(map[string]time.Time)
	mu       sync.Mutex
)

func InitAuth() {
	authUsername = strings.TrimSpace(os.Getenv("user"))
	authPassword = strings.TrimSpace(os.Getenv("password"))

	if authUsername == "" || authPassword == "" {
		log.Fatal("ERRO FATAL: Credenciais de autenticação (user, password) não definidas no ambiente.")
	}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isValidSession(token string) bool {
	mu.Lock()
	defer mu.Unlock()
	
	exp, exists := sessions[token]
	if !exists {
		return false
	}
	
	if time.Now().After(exp) {
		delete(sessions, token)
		return false
	}
	
	return true
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.APIResponse{Error: message})
}

func sendJSONSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Message: message, Data: data})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	if req.Username == authUsername && req.Password == authPassword {
		token := generateToken()
		
		mu.Lock()
		sessions[token] = time.Now().Add(24 * time.Hour)
		mu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "alfredo_session",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})

		sendJSONSuccess(w, "Login bem-sucedido", nil)
		return
	}

	sendJSONError(w, "Credenciais inválidas.", http.StatusUnauthorized)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("alfredo_session")
	if err == nil {
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "alfredo_session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	sendJSONSuccess(w, "Logout bem-sucedido", nil)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("alfredo_session")
		
		if err != nil || !isValidSession(cookie.Value) {
			sendJSONError(w, "Não autorizado. Faça o login.", http.StatusUnauthorized)
			log.Printf("Tentativa de acesso NÃO AUTORIZADO de %s para %s %s", r.RemoteAddr, r.Method, r.URL.Path)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Requisição Recebida: [IP: %s] [Método: %s] [URL: %s]", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Requisição Concluída: [IP: %s] [Método: %s] [URL: %s] - Tempo: %s", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
