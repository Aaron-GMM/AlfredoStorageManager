package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/joho/godotenv"

	"alfredostoragemanager/api/handlers"
	"alfredostoragemanager/api/middleware"
	"alfredostoragemanager/api/services"
)

func handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		ApiBaseUrl string `json:"api_base_url"`
	}{
		// Agora o frontend e a API estão no mesmo host
		ApiBaseUrl: "", 
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não encontrado ou erro ao carregar. Usando variáveis de ambiente do sistema.")
	}

	// Initialize Authentication from environment
	middleware.InitAuth()

	// Initialize Storage Service
	basePath := os.Getenv("BASE_DIR_PATH")
	if basePath == "" {
		log.Fatal("ERRO FATAL: BASE_DIR_PATH não definida nas variáveis de ambiente.")
	}

	storageService, err := services.NewLocalDiskStorage(basePath)
	if err != nil {
		log.Fatalf("ERRO FATAL: Falha ao inicializar o serviço de storage: %v", err)
	}

	// Initialize Handlers
	storageHandler := handlers.NewStorageHandler(storageService)

	// Setup Routes and Middleware
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/config", handleAPIConfig)
	mux.HandleFunc("/api/login", middleware.HandleLogin)
	mux.HandleFunc("/api/logout", middleware.HandleLogout)

	// Protected API Routes
	mux.Handle("/files", middleware.Auth(http.HandlerFunc(storageHandler.HandleListFiles)))
	mux.Handle("/create-folder", middleware.Auth(http.HandlerFunc(storageHandler.HandleCreateFolder)))
	mux.Handle("/delete", middleware.Auth(http.HandlerFunc(storageHandler.HandleDelete)))
	mux.Handle("/rename", middleware.Auth(http.HandlerFunc(storageHandler.HandleRename)))
	mux.Handle("/download", middleware.Auth(http.HandlerFunc(storageHandler.HandleDownload)))
	mux.Handle("/upload", middleware.Auth(http.HandlerFunc(storageHandler.HandleUpload)))

	// Pprof Routes (Profiling Ativo de Memória e CPU)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Apply global middleware: Logging -> CORS -> Mux
	handler := middleware.Logging(middleware.CORS(mux))

	port := "8080"
	
	// Server com timeouts estritos para evitar vazamento de goroutines (Concorrência Limpa)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,  // Tempo máximo lendo a requisição
		WriteTimeout: 60 * time.Second,  // Tempo máximo escrevendo a resposta (uploads podem demorar um pouco mais se houver lentidão na rede, ajustar conforme a necessidade de stream)
		IdleTimeout:  120 * time.Second, // Tempo máximo mantendo a conexão aberta
	}

	log.Printf("Servidor Backend Go (com Frontend Embutido) iniciado na porta :%s\n", port)
	log.Fatal(server.ListenAndServe())
}