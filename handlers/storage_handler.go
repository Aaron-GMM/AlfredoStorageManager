package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"alfredostoragemanager/api/models"
	"alfredostoragemanager/api/services"
)

type StorageHandler struct {
	storage services.StorageService
}

func NewStorageHandler(storage services.StorageService) *StorageHandler {
	return &StorageHandler{storage: storage}
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

func (h *StorageHandler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Query().Get("path")

	items, currentPath, err := h.storage.ListDir(reqPath)
	if err != nil {
		if err == services.ErrAccessDenied {
			sendJSONError(w, "Acesso restrito. Tentativa de navegar fora da área autorizada.", http.StatusForbidden)
			return
		}
		if err == services.ErrNotFound {
			sendJSONError(w, "Caminho não encontrado ou sem acesso.", http.StatusNotFound)
			return
		}
		if err == services.ErrNotADirectory {
			sendJSONError(w, "O caminho especificado não é uma pasta.", http.StatusBadRequest)
			return
		}
		sendJSONError(w, "Erro interno ao listar arquivos.", http.StatusInternalServerError)
		return
	}

	response := models.FilesResponse{
		CurrentPath: currentPath,
		AppRootPath: h.storage.GetBasePath(),
		Files:       items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StorageHandler) HandleCreateFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	if err := h.storage.CreateFolder(req.Path, req.FolderName); err != nil {
		if err == services.ErrAccessDenied {
			sendJSONError(w, "Acesso restrito.", http.StatusForbidden)
			return
		}
		sendJSONError(w, "Falha ao criar pasta.", http.StatusInternalServerError)
		return
	}

	sendJSONSuccess(w, fmt.Sprintf("Pasta '%s' criada com sucesso.", req.FolderName), nil)
}

func (h *StorageHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		sendJSONError(w, "Caminho não informado.", http.StatusBadRequest)
		return
	}

	if err := h.storage.Delete(reqPath); err != nil {
		if err == services.ErrAccessDenied {
			sendJSONError(w, "Acesso restrito.", http.StatusForbidden)
			return
		}
		sendJSONError(w, "Falha ao deletar item.", http.StatusInternalServerError)
		return
	}

	sendJSONSuccess(w, "Item deletado com sucesso.", nil)
}

func (h *StorageHandler) HandleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	var req models.RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Requisição inválida.", http.StatusBadRequest)
		return
	}

	if err := h.storage.Rename(req.OldPath, req.NewName); err != nil {
		if err == services.ErrAccessDenied {
			sendJSONError(w, "Acesso restrito.", http.StatusForbidden)
			return
		}
		sendJSONError(w, "Falha ao renomear item.", http.StatusInternalServerError)
		return
	}

	sendJSONSuccess(w, "Item renomeado com sucesso.", nil)
}

func (h *StorageHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	reqPath := r.URL.Query().Get("path")
	securePath, err := h.storage.GetSecurePath(reqPath)
	if err != nil {
		sendJSONError(w, "Acesso restrito.", http.StatusForbidden)
		return
	}

	// http.ServeFile effectively handles Range headers and streaming,
	// which prevents high RAM consumption on large files.
	http.ServeFile(w, r, securePath)
}

func (h *StorageHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	targetDir := r.URL.Query().Get("path")
	if targetDir == "" {
		sendJSONError(w, "Caminho de destino não especificado.", http.StatusBadRequest)
		return
	}

	// Using MultipartReader instead of ParseMultipartForm
	// to stream the upload efficiently without loading chunks to RAM/TempFiles
	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("Erro ao iniciar MultipartReader: %v", err)
		sendJSONError(w, "Falha ao ler multipart request.", http.StatusBadRequest)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler parte do upload: %v", err)
			sendJSONError(w, "Erro na leitura do upload.", http.StatusInternalServerError)
			return
		}

		filename := part.FileName()
		if filename == "" {
			continue // Skip non-file parts
		}

		// Clean filename to avoid directory traversal via filename (e.g., ../file.txt)
		cleanFileName := filepath.Base(filename)

		err = h.storage.SaveStream(targetDir, cleanFileName, part)
		if err != nil {
			log.Printf("Erro ao salvar arquivo %s: %v", cleanFileName, err)
			sendJSONError(w, "Erro ao salvar o arquivo no disco.", http.StatusInternalServerError)
			return
		}
	}

	sendJSONSuccess(w, "Upload(s) concluído(s) com sucesso.", nil)
}
