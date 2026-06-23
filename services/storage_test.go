package services

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkListDir(b *testing.B) {
	// Preparar um diretório temporário com alguns arquivos para o benchmark
	tempDir := b.TempDir()
	storage, err := NewLocalDiskStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}

	// Criar alguns arquivos dummy
	for i := 0; i < 100; i++ {
		filePath := filepath.Join(tempDir, "file_"+string(rune(i))+".txt")
		_ = os.WriteFile(filePath, []byte("dummy data"), 0644)
	}

	// Resetar o timer de benchmark para ignorar o tempo de setup
	b.ResetTimer()

	// O desenvolvedor deve rodar `go test -bench=. -benchmem`
	for i := 0; i < b.N; i++ {
		_, _, err := storage.ListDir("")
		if err != nil {
			b.Fatalf("ListDir failed: %v", err)
		}
	}
}

func BenchmarkGetSecurePath(b *testing.B) {
	tempDir := b.TempDir()
	storage, err := NewLocalDiskStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}

	testPath := "some/test/path"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetSecurePath(testPath)
	}
}
