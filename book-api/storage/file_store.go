package storage

import (
	"book-api/models"
	"encoding/json"
	"os"
	"sync"
)

type FileStore struct {
	filePath string
	mu       sync.RWMutex
}

func NewFileStore(filePath string) *FileStore {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte("[]"), 0644)
	}
	return &FileStore{filePath: filePath}
}

func (fs *FileStore) ReadAll() ([]*models.Book, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var books []*models.Book
	if len(data) == 0 {
		return books, nil
	}

	err = json.Unmarshal(data, &books)
	return books, err
}

func (fs *FileStore) WriteAll(books []*models.Book) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := fs.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fs.filePath)
}
