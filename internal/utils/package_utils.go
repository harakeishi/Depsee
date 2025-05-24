package utils

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	// 標準ライブラリのキャッシュ
	standardLibsCache map[string]bool
	standardLibsOnce  sync.Once
)

// IsStandardLibrary は指定されたimportパスが標準ライブラリかどうかを判定
func IsStandardLibrary(importPath string) bool {
	// 空文字列やドットのみのパスは標準ライブラリではない
	if importPath == "" || importPath == "." {
		return false
	}

	// ドメイン名を含むパッケージ（例：github.com/user/repo）は外部パッケージ
	if strings.Contains(importPath, ".") {
		return false
	}

	// スラッシュを含まない単一パッケージ名は標準ライブラリの可能性が高い
	if !strings.Contains(importPath, "/") {
		return isKnownStandardLibrary(importPath)
	}

	// パスの最初の部分が標準ライブラリかチェック
	parts := strings.Split(importPath, "/")
	if len(parts) > 0 {
		return isKnownStandardLibrary(parts[0])
	}

	return false
}

// isKnownStandardLibrary は動的に取得した標準ライブラリリストまたはフォールバックリストを使用して判定
func isKnownStandardLibrary(pkgName string) bool {
	standardLibsOnce.Do(func() {
		standardLibsCache = getStandardLibraries()
	})

	return standardLibsCache[pkgName]
}

// getStandardLibraries は標準ライブラリの一覧を取得（動的取得 + フォールバック）
func getStandardLibraries() map[string]bool {
	// まず動的に取得を試行
	if stdLibs := getStandardLibrariesDynamic(); stdLibs != nil {
		return stdLibs
	}

	// 動的取得に失敗した場合はフォールバックリストを使用
	return getStandardLibrariesFallback()
}

// getStandardLibrariesDynamic は`go list std`コマンドを使用して動的に標準ライブラリを取得
func getStandardLibrariesDynamic() map[string]bool {
	cmd := exec.Command("go", "list", "std")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	stdLibs := make(map[string]bool)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// パッケージ名とパスの最初の部分を両方登録
		stdLibs[line] = true

		// パスの最初の部分も登録（例：crypto/sha256 -> crypto）
		if idx := strings.Index(line, "/"); idx != -1 {
			stdLibs[line[:idx]] = true
		}
	}

	return stdLibs
}

// getStandardLibrariesFallback はハードコードされたフォールバックリストを返す
func getStandardLibrariesFallback() map[string]bool {
	// Go標準ライブラリの主要パッケージ
	// 参考: https://pkg.go.dev/std
	return map[string]bool{
		// Core packages
		"builtin": true, "unsafe": true,

		// Common packages
		"fmt": true, "os": true, "io": true, "time": true, "strings": true, "strconv": true,
		"context": true, "sync": true, "errors": true, "sort": true, "math": true,
		"bytes": true, "bufio": true, "path": true, "reflect": true, "regexp": true,
		"runtime": true, "syscall": true, "testing": true, "unicode": true,

		// Network and HTTP
		"net": true, "http": true,

		// Encoding and crypto
		"encoding": true, "crypto": true, "hash": true,

		// Compression and containers
		"compress": true, "container": true,

		// Development and debugging
		"debug": true, "expvar": true, "flag": true, "log": true,

		// File and path handling
		"filepath": true, "mime": true,

		// HTML, image, and text processing
		"html": true, "image": true, "text": true,

		// Database and indexing
		"database": true, "index": true,

		// Go toolchain
		"go": true,

		// Archive
		"archive": true,

		// Additional common packages
		"cmp": true, "slices": true, "maps": true,
	}
}

// ExtractPackageName はimportパスからパッケージ名を抽出
func ExtractPackageName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

// IsLocalPackage は指定されたimportパスが同リポジトリ内のパッケージかどうかを判定
func IsLocalPackage(importPath string) bool {
	// 相対パスは常にローカルパッケージ
	if strings.HasPrefix(importPath, ".") {
		return true
	}

	// 標準ライブラリの判定を改善
	return !IsStandardLibrary(importPath)
}
