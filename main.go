package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)


func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

type FileItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
}


type Server struct {
	basePath string
}


func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Olá do meu servidor Go no Windows!")
}


func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Query().Get("path")

	currentDir := s.basePath 
	if reqPath != "" {
		resolvedPath := filepath.Join(s.basePath, reqPath) 

		if !strings.HasPrefix(resolvedPath, s.basePath) { 
			http.Error(w, "Acesso negado: Tentativa de acessar fora do diretório base", http.StatusForbidden)
			return
		}
		currentDir = resolvedPath
	}

	fileInfo, err := os.Stat(currentDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro: O caminho '%s' não existe ou não pode ser acessado: %v", currentDir, err), http.StatusNotFound)
		return
	}
	if !fileInfo.IsDir() {
		http.Error(w, fmt.Sprintf("Erro: '%s' não é um diretório", currentDir), http.StatusBadRequest)
		return
	}

	files, err := os.ReadDir(currentDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao listar conteúdo de '%s': %v", currentDir, err), http.StatusInternalServerError)
		return
	}

	var fileItems []FileItem
	for _, file := range files {
		fileItems = append(fileItems, FileItem{
			Name:  file.Name(),
			IsDir: file.IsDir(),
		})
	}

	w.Header().Set("Content-Type", "application/json")

	displayPath := strings.TrimPrefix(currentDir, s.basePath) 
	if displayPath == "" {
		displayPath = s.basePath 
	}

	response := struct {
		CurrentPath string     `json:"current_path"`
		Files       []FileItem `json:"files"`
	}{
		CurrentPath: displayPath,
		Files:       fileItems,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erro ao codificar resposta JSON", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		Path       string `json:"path"`
		FolderName string `json:"folder_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Erro ao decodificar JSON da requisição", http.StatusBadRequest)
		return
	}

	var targetPath string
	
	if requestBody.Path == strings.TrimSuffix(s.basePath, string(os.PathSeparator)) || requestBody.Path == s.basePath {
		targetPath = filepath.Join(s.basePath, requestBody.FolderName) 
	} else {
		targetPath = filepath.Join(s.basePath, requestBody.Path, requestBody.FolderName) 
	}

	if !strings.HasPrefix(targetPath, s.basePath) { 
		http.Error(w, "Acesso negado: Tentativa de criar pasta fora do diretório base", http.StatusForbidden)
		return
	}

	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao criar pasta '%s': %v", requestBody.FolderName, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pasta '%s' criada com sucesso em '%s'", requestBody.FolderName, requestBody.Path)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: .env não encontrado ou erro ao carregar. As variáveis de ambiente serão lidas diretamente do sistema.")
	}

	basePathValue := os.Getenv("BASE_DIR_PATH")
	if basePathValue == "" {
		log.Println("BASE_DIR_PATH não definida nas variáveis de ambiente. Usando 'S:\\' como padrão.")
		basePathValue = "S:\\"
	}

	basePathValue = filepath.Clean(basePathValue)
	if !strings.HasSuffix(basePathValue, string(os.PathSeparator)) {
		basePathValue += string(os.PathSeparator)
	}

	s := &Server{
		basePath: basePathValue,
	}

	http.HandleFunc("/", enableCORS(s.handleRoot))
	http.HandleFunc("/files", enableCORS(s.handleFiles))
	http.HandleFunc("/create-folder", enableCORS(s.handleCreateFolder))

	port := "8080"
	log.Printf("Servidor Backend Go iniciado localmente na porta :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}