package storage

import (
	"fmt"
	"sync"
)

// MemoryStorage はメモリ上でのストレージ実装
type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewMemoryStorage は新しいメモリストレージを作成
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

// Read はキーに対応する値を読み取り
func (m *MemoryStorage) Read(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	value, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// ReadAll は全てのデータを読み取り
func (m *MemoryStorage) ReadAll() (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]string)
	for k, v := range m.data {
		result[k] = v
	}
	return result, nil
}

// Write はキーと値を書き込み
func (m *MemoryStorage) Write(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.data[key] = value
	return nil
}

// Delete はキーを削除
func (m *MemoryStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.data, key)
	return nil
}
