package models

type FileItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
}

type APIResponse struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type FilesResponse struct {
	CurrentPath string     `json:"current_path"`
	AppRootPath string     `json:"app_root_path"`
	Files       []FileItem `json:"files"`
}

type CreateFolderRequest struct {
	Path       string `json:"path"`
	FolderName string `json:"folder_name"`
}

type RenameRequest struct {
	OldPath string `json:"old_path"`
	NewName string `json:"new_name"`
}
