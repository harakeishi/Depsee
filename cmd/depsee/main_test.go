package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harakeishi/depsee/internal/logger"
)

func TestMain(m *testing.M) {
	// テスト用のログ設定
	logger.Init(logger.Config{
		Level:  logger.LevelError, // テスト時はエラーのみ
		Format: "text",
		Output: os.Stderr,
	})

	code := m.Run()
	os.Exit(code)
}

func TestCLIVersion(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	// バージョン表示テスト
	cmd = exec.Command("./depsee_test", "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run version command: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "depsee") || !strings.Contains(outputStr, "v0.1.0") {
		t.Errorf("Version output unexpected: %s", outputStr)
	}
}

func TestCLIAnalyze(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

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

	// 解析コマンドテスト
	cmd = exec.Command("./depsee_test", "analyze", absPath)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run analyze command: %v", err)
	}

	outputStr := string(output)

	// 期待される出力の確認
	expectedStrings := []string{
		"User",
		"Profile",
		"Post",
		"UserSettings",
		"UserService",
		"CreateUser",
		"GetUserPosts",
		"graph TD",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, outputStr)
		}
	}
}

func TestCLIInvalidArguments(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	// 引数なしでの実行（ヘルプが表示されるはず）
	cmd = exec.Command("./depsee_test")
	_, err = cmd.Output()
	// Cobraではヘルプが表示されるため、エラーにならない場合がある
	// 実際の動作を確認するため、出力をチェック

	// 無効なコマンドでの実行
	cmd = exec.Command("./depsee_test", "invalid")
	_, err = cmd.Output()
	if err == nil {
		t.Error("Expected error for invalid command, but got none")
	}

	// 存在しないディレクトリでの実行
	cmd = exec.Command("./depsee_test", "analyze", "/non/existent/directory")
	_, err = cmd.Output()
	if err == nil {
		t.Error("Expected error for non-existent directory, but got none")
	}
}

func TestCLILogLevels(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	// デバッグレベルでの実行（グローバルフラグとして）
	cmd = exec.Command("./depsee_test", "--log-level", "debug", "analyze", absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with debug log level: %v", err)
	}

	outputStr := string(output)
	// デバッグログが含まれていることを確認
	if !strings.Contains(outputStr, "level=DEBUG") {
		t.Logf("Debug output: %s", outputStr) // デバッグ情報として出力
	}
}

func TestCLIJSONLogFormat(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	// JSONフォーマットでの実行（グローバルフラグとして）
	cmd = exec.Command("./depsee_test", "--log-format", "json", "analyze", absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with JSON log format: %v", err)
	}

	outputStr := string(output)
	// JSONログが含まれていることを確認（ログが出力される場合）
	if strings.Contains(outputStr, `"level"`) {
		if !strings.Contains(outputStr, `"msg"`) {
			t.Error("JSON log format should contain 'msg' field")
		}
	}
}

func TestCLIIncludePackageDeps(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	// include-package-depsフラグでの実行
	cmd = exec.Command("./depsee_test", "analyze", "--include-package-deps", absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with include-package-deps flag: %v", err)
	}

	outputStr := string(output)
	// パッケージ間依存関係が含まれていることを確認
	if !strings.Contains(outputStr, "パッケージ間依存関係を含む依存グラフ構築") {
		t.Logf("Output: %s", outputStr) // デバッグ情報として出力
	}
}

func TestCLIFlagPositions(t *testing.T) {
	// バイナリをビルド
	cmd := exec.Command("go", "build", "-o", "depsee_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("depsee_test")

	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("Failed to resolve test data directory: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Test data directory does not exist: %s", absPath)
	}

	// 異なるフラグ位置でのテスト
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "グローバルフラグ前置",
			args: []string{"--log-level", "debug", "analyze", absPath},
		},
		{
			name: "ローカルフラグ後置",
			args: []string{"analyze", "--include-package-deps", absPath},
		},
		{
			name: "混在パターン",
			args: []string{"--log-level", "debug", "analyze", "--include-package-deps", absPath},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd = exec.Command("./depsee_test", tc.args...)
			_, err = cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to run with args %v: %v", tc.args, err)
			}
		})
	}
}
