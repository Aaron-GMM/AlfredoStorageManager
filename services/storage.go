package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"alfredostoragemanager/api/models"
)

var (
	ErrAccessDenied = errors.New("access denied: path traversal attempt")
	ErrNotFound     = errors.New("path not found or inaccessible")
	ErrNotADirectory = errors.New("path is not a directory")
	ErrInternal     = errors.New("internal server error")
)

type StorageService interface {
	ListDir(targetPath string) ([]models.FileItem, string, error)
	CreateFolder(parentPath string, folderName string) error
	Delete(targetPath string) error
	Rename(oldPath string, newName string) error
	GetSecurePath(targetPath string) (string, error)
	GetBasePath() string
	SaveStream(targetDir string, filename string, reader io.Reader) error
}

type LocalDiskStorage struct {
	basePath string
}

func NewLocalDiskStorage(basePath string) (*LocalDiskStorage, error) {
	cleanPath := filepath.Clean(basePath)
	if !strings.HasSuffix(cleanPath, string(os.PathSeparator)) {
		cleanPath += string(os.PathSeparator)
	}

	info, err := os.Stat(cleanPath)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("base path %s is not a valid directory", cleanPath)
	}

	return &LocalDiskStorage{basePath: cleanPath}, nil
}

func (s *LocalDiskStorage) GetBasePath() string {
	return s.basePath
}

// GetSecurePath validates that a requested path is within the base path
func (s *LocalDiskStorage) GetSecurePath(targetPath string) (string, error) {
	if targetPath == "" {
		targetPath = s.basePath
	}

	cleanedTarget := filepath.Clean(targetPath)
	
	// Add separator for correct prefix comparison (e.g. /var/www vs /var/www2)
	checkTarget := cleanedTarget
	if !strings.HasSuffix(checkTarget, string(os.PathSeparator)) {
		checkTarget += string(os.PathSeparator)
	}

	if !strings.HasPrefix(checkTarget, s.basePath) && cleanedTarget != strings.TrimSuffix(s.basePath, string(os.PathSeparator)) {
		return "", ErrAccessDenied
	}

	return cleanedTarget, nil
}

func (s *LocalDiskStorage) ListDir(targetPath string) ([]models.FileItem, string, error) {
	securePath, err := s.GetSecurePath(targetPath)
	if err != nil {
		return nil, "", err
	}

	fileInfo, err := os.Stat(securePath)
	if err != nil {
		return nil, "", ErrNotFound
	}

	if !fileInfo.IsDir() {
		return nil, "", ErrNotADirectory
	}

	entries, err := os.ReadDir(securePath)
	if err != nil {
		return nil, "", ErrInternal
	}

	var items []models.FileItem
	for _, entry := range entries {
		info, err := entry.Info()
		size := int64(0)
		if err == nil {
			size = info.Size()
		}

		items = append(items, models.FileItem{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  size,
			Path:  filepath.Join(securePath, entry.Name()),
		})
	}

	return items, securePath, nil
}

func (s *LocalDiskStorage) CreateFolder(parentPath string, folderName string) error {
	secureParent, err := s.GetSecurePath(parentPath)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(secureParent, folderName)
	secureTarget, err := s.GetSecurePath(targetPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(secureTarget, 0755); err != nil {
		return ErrInternal
	}

	return nil
}

func (s *LocalDiskStorage) Delete(targetPath string) error {
	securePath, err := s.GetSecurePath(targetPath)
	if err != nil {
		return err
	}

	if securePath == strings.TrimSuffix(s.basePath, string(os.PathSeparator)) {
		return errors.New("cannot delete root base path")
	}

	if err := os.RemoveAll(securePath); err != nil {
		return ErrInternal
	}

	return nil
}

func (s *LocalDiskStorage) Rename(oldPath string, newName string) error {
	secureOldPath, err := s.GetSecurePath(oldPath)
	if err != nil {
		return err
	}

	if secureOldPath == strings.TrimSuffix(s.basePath, string(os.PathSeparator)) {
		return errors.New("cannot rename root base path")
	}

	parentDir := filepath.Dir(secureOldPath)
	secureNewPath, err := s.GetSecurePath(filepath.Join(parentDir, newName))
	if err != nil {
		return err
	}

	if err := os.Rename(secureOldPath, secureNewPath); err != nil {
		return ErrInternal
	}

	return nil
}

func (s *LocalDiskStorage) SaveStream(targetDir string, filename string, reader io.Reader) error {
	secureDir, err := s.GetSecurePath(targetDir)
	if err != nil {
		return err
	}

	secureFilePath, err := s.GetSecurePath(filepath.Join(secureDir, filename))
	if err != nil {
		return err
	}

	outFile, err := os.Create(secureFilePath)
	if err != nil {
		return ErrInternal
	}
	defer outFile.Close()

	// io.Copy uses a small buffer (32KB default) to stream from reader to file.
	// Highly memory-efficient for 1GB RAM constraint.
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	return nil
}
