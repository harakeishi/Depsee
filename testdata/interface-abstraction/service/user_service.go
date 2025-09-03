package service

import (
	"github.com/harakeishi/depsee/testdata/interface-abstraction/storage"
)

// SimpleService はシンプルなサービス例
type SimpleService struct {
	storage storage.Storage // インターフェースに依存
	logger  storage.Logger  // インターフェースに依存
}

// NewSimpleService は新しいサービスを作成（依存性注入）
func NewSimpleService(storage storage.Storage, logger storage.Logger) *SimpleService {
	return &SimpleService{
		storage: storage, // Storage interface <- 具象実装が注入される
		logger:  logger,  // Logger interface <- 具象実装が注入される
	}
}

// Process はデータを処理
func (s *SimpleService) Process(key, value string) {
	s.logger.Log("Processing data")        // Logger interface -> 具象実装のメソッド
	s.storage.Write(key, value)           // Storage interface -> 具象実装のメソッド
	s.logger.Log("Processing completed")  // Logger interface -> 具象実装のメソッド
}
