package service

import (
	"github.com/harakeishi/depsee/testdata/interface-abstraction/storage"
)

// App はアプリケーションのメイン構造体
type App struct {
	storage storage.Storage // インターフェースに依存
	logger  storage.Logger  // インターフェースに依存
}

// NewApp は新しいアプリケーションを作成（依存性注入）
func NewApp(storage storage.Storage, logger storage.Logger) *App {
	return &App{
		storage: storage, // Storage interface <- 具象実装が注入される
		logger:  logger,  // Logger interface <- 具象実装が注入される
	}
}

// Run はアプリケーションを実行
func (a *App) Run() {
	a.logger.Log("Application starting...")
	a.storage.Write("status", "running")
	a.logger.Log("Application finished.")
}

// CreateWithMemoryAndConsole はメモリ+コンソールの組み合わせ
func CreateWithMemoryAndConsole() *App {
	memStorage := storage.NewMemoryStorage()    // 具象: MemoryStorage
	consoleLogger := storage.NewConsoleLogger() // 具象: ConsoleLogger
	return NewApp(memStorage, consoleLogger)    // 注入: Storage <- MemoryStorage, Logger <- ConsoleLogger
}

// CreateWithFileAndConsole はファイル+コンソールの組み合わせ
func CreateWithFileAndConsole(path string) *App {
	fileStorage := storage.NewFileStorage(path) // 具象: FileStorage
	consoleLogger := storage.NewConsoleLogger() // 具象: ConsoleLogger
	return NewApp(fileStorage, consoleLogger)   // 注入: Storage <- FileStorage, Logger <- ConsoleLogger
}

// CreateWithFileAndFileLogger はファイル+ファイルログの組み合わせ
func CreateWithFileAndFileLogger(dataPath, logPath string) *App {
	fileStorage := storage.NewFileStorage(dataPath) // 具象: FileStorage (データ用)
	logStorage := storage.NewFileStorage(logPath)   // 具象: FileStorage (ログ用)
	fileLogger := storage.NewFileLogger(logStorage) // 具象: FileLogger <- FileStorage
	return NewApp(fileStorage, fileLogger)          // 注入: Storage <- FileStorage, Logger <- FileLogger
}
