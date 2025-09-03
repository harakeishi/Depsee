package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileStorage はファイルシステムでのストレージ実装
type FileStorage struct {
	basePath string
}

// NewFileStorage は新しいファイルストレージを作成
func NewFileStorage(basePath string) Storage {
	return &FileStorage{
		basePath: basePath,
	}
}

// Read はファイルからデータを読み取り
func (f *FileStorage) Read(key string) (string, error) {
	filePath := filepath.Join(f.basePath, key+".txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", key, err)
	}
	return string(data), nil
}

// ReadAll は全てのファイルからデータを読み取り
func (f *FileStorage) ReadAll() (map[string]string, error) {
	result := make(map[string]string)
	
	entries, err := os.ReadDir(f.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			key := strings.TrimSuffix(entry.Name(), ".txt")
			value, err := f.Read(key)
			if err != nil {
				continue
			}
			result[key] = value
		}
	}
	
	return result, nil
}

// Write はファイルにデータを書き込み
func (f *FileStorage) Write(key, value string) error {
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(f.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	filePath := filepath.Join(f.basePath, key+".txt")
	err := os.WriteFile(filePath, []byte(value), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", key, err)
	}
	return nil
}

// Delete はファイルを削除
func (f *FileStorage) Delete(key string) error {
	filePath := filepath.Join(f.basePath, key+".txt")
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %v", key, err)
	}
	return nil
}
