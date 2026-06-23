package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"alfredostoragemanager/api/handlers"
	"alfredostoragemanager/api/middleware"
	"alfredostoragemanager/api/services"
)

//go:embed Frontend
var frontendFiles embed.FS

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

	// Extrair o sub-sistema de arquivos da pasta embutida "Frontend"
	frontendFS, err := fs.Sub(frontendFiles, "Frontend")
	if err != nil {
		log.Fatalf("ERRO FATAL: Falha ao carregar o frontend embutido: %v", err)
	}

	// Setup Routes and Middleware
	mux := http.NewServeMux()

	// Servir os arquivos estáticos do Frontend embutido na raiz "/"
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))

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

	// Apply global middleware: Logging -> CORS -> Mux
	handler := middleware.Logging(middleware.CORS(mux))

	port := "8080"
	log.Printf("Servidor Backend Go (com Frontend Embutido) iniciado na porta :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}