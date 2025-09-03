package storage

import (
	"fmt"
	"time"
)

// ConsoleLogger はコンソール出力ロガーの実装
type ConsoleLogger struct{}

// NewConsoleLogger は新しいコンソールロガーを作成
func NewConsoleLogger() Logger {
	return &ConsoleLogger{}
}

// Log は通常のログメッセージを出力
func (c *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Error はエラーメッセージを出力
func (c *ConsoleLogger) Error(message string) {
	fmt.Printf("[%s] ERROR: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// FileLogger はファイル出力ロガーの実装
type FileLogger struct {
	storage Storage
}

// NewFileLogger は新しいファイルロガーを作成
func NewFileLogger(storage Storage) Logger {
	return &FileLogger{
		storage: storage,
	}
}

// Log は通常のログメッセージをファイルに出力
func (f *FileLogger) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] INFO: %s", timestamp, message)
	
	// ログをファイルに追記
	existing, _ := f.storage.Read("log")
	newLog := existing + logMessage + "\n"
	f.storage.Write("log", newLog)
}

// Error はエラーメッセージをファイルに出力
func (f *FileLogger) Error(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] ERROR: %s", timestamp, message)
	
	// エラーログをファイルに追記
	existing, _ := f.storage.Read("error")
	newLog := existing + logMessage + "\n"
	f.storage.Write("error", newLog)
}
