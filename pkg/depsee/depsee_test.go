package depsee

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/output"
)

func TestNew(t *testing.T) {
	app := New()
	if app == nil {
		t.Error("New() should return a non-nil Depsee instance")
	}
}

func TestNewWithDependencies(t *testing.T) {
	analyzer := analyzer.New()
	grapher := graph.NewBuilder()
	outputter := output.NewGenerator()
	logger := logger.NewLogger(logger.Config{
		Level:  logger.LevelError,
		Format: "text",
		Output: os.Stderr,
	})

	app := NewWithDependencies(analyzer, grapher, outputter, logger)
	if app == nil {
		t.Error("NewWithDependencies() should return a non-nil Depsee instance")
	}
}

func TestAnalyze_NonExistentDirectory(t *testing.T) {
	app := New()
	config := Config{
		TargetDir:          "/non/existent/directory",
		IncludePackageDeps: false,
		LogLevel:           "error",
		LogFormat:          "text",
	}

	err := app.Analyze(config)
	if err == nil {
		t.Error("Analyze() should return an error for non-existent directory")
	}
}

func TestAnalyze_ValidDirectory(t *testing.T) {
	// テストデータディレクトリのパス
	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	app := New()
	config := Config{
		TargetDir:          absPath,
		IncludePackageDeps: false,
		LogLevel:           "error", // テスト時はエラーのみ
		LogFormat:          "text",
	}

	err = app.Analyze(config)
	if err != nil {
		t.Errorf("Analyze() should not return an error for valid directory: %v", err)
	}
}

func TestAnalyze_WithPackageDeps(t *testing.T) {
	// テストデータディレクトリのパス
	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	app := New()
	config := Config{
		TargetDir:          absPath,
		IncludePackageDeps: true,
		LogLevel:           "error", // テスト時はエラーのみ
		LogFormat:          "text",
	}

	err = app.Analyze(config)
	if err != nil {
		t.Errorf("Analyze() should not return an error with package deps: %v", err)
	}
}

func TestConfig(t *testing.T) {
	config := Config{
		TargetDir:          "/some/path",
		IncludePackageDeps: true,
		LogLevel:           "debug",
		LogFormat:          "json",
	}

	if config.TargetDir != "/some/path" {
		t.Errorf("Expected TargetDir to be '/some/path', got '%s'", config.TargetDir)
	}

	if !config.IncludePackageDeps {
		t.Error("Expected IncludePackageDeps to be true")
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", config.LogLevel)
	}

	if config.LogFormat != "json" {
		t.Errorf("Expected LogFormat to be 'json', got '%s'", config.LogFormat)
	}
}
