package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	var targetDir string
	if reqPath == "" {

		targetDir = s.basePath
	} else {

		targetDir = reqPath
	}

	cleanedTargetDir := filepath.Clean(targetDir)
	cleanedBasePath := filepath.Clean(s.basePath)


	if !strings.HasSuffix(cleanedBasePath, string(os.PathSeparator)) {
		cleanedBasePath += string(os.PathSeparator)
	}


	if cleanedTargetDir != string(os.PathSeparator) && !strings.HasSuffix(cleanedTargetDir, string(os.PathSeparator)) && !(len(cleanedTargetDir) == 2 && cleanedTargetDir[1] == ':') {
		cleanedTargetDir += string(os.PathSeparator)
	}
	
	if len(cleanedTargetDir) == 2 && cleanedTargetDir[1] == ':' && !strings.HasSuffix(cleanedTargetDir, string(os.PathSeparator)) {
		cleanedTargetDir += string(os.PathSeparator)
	}
	log.Printf("DEBUG [handleFiles]: s.basePath (servidor): '%s'", s.basePath)
	log.Printf("DEBUG [handleFiles]: reqPath (do frontend): '%s'", reqPath)
	log.Printf("DEBUG [handleFiles]: cleanedBasePath (para comparação): '%s'", cleanedBasePath)
	log.Printf("DEBUG [handleFiles]: cleanedTargetDir (para comparação): '%s'", cleanedTargetDir)

	if !strings.HasPrefix(cleanedTargetDir, cleanedBasePath) {
		log.Printf("ERRO DE SEGURANÇA [handleFiles]: '%s' NÃO começa com '%s'. Acesso negado.", cleanedTargetDir, cleanedBasePath)
		http.Error(w, "Acesso negado: Tentativa de acessar fora do diretório base", http.StatusForbidden)
		return
	}

	currentDir := cleanedTargetDir

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

	displayPath := currentDir

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

	cleanedBasePath := filepath.Clean(s.basePath)
	if !strings.HasSuffix(cleanedBasePath, string(os.PathSeparator)) {
		cleanedBasePath += string(os.PathSeparator)
	}

	targetPath := filepath.Join(requestBody.Path, requestBody.FolderName)

	cleanedTargetPath := filepath.Clean(targetPath)
	if len(cleanedTargetPath) == 2 && cleanedTargetPath[1] == ':' && !strings.HasSuffix(cleanedTargetPath, string(os.PathSeparator)) {
		cleanedTargetPath += string(os.PathSeparator)
	}
	if cleanedTargetPath != string(os.PathSeparator) && !strings.HasSuffix(cleanedTargetPath, string(os.PathSeparator)) && !(len(cleanedTargetPath) == 2 && cleanedTargetPath[1] == ':') {
		cleanedTargetPath += string(os.PathSeparator)
	}
	if len(cleanedTargetPath) == 2 && cleanedTargetPath[1] == ':' && !strings.HasSuffix(cleanedTargetPath, string(os.PathSeparator)) {
		cleanedTargetPath += string(os.PathSeparator)
	}
	if !strings.HasPrefix(cleanedTargetPath, cleanedBasePath) {
		http.Error(w, "Acesso negado: Tentativa de criar pasta fora do diretório base", http.StatusForbidden)
		return
	}

	err := os.MkdirAll(cleanedTargetPath, 0755)
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