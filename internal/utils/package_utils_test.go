package utils

import (
	"testing"
)

func TestIsStandardLibrary(t *testing.T) {
	tests := []struct {
		importPath string
		expected   bool
	}{
		// 標準ライブラリ
		{"fmt", true},
		{"time", true},
		{"net/http", true},
		{"encoding/json", true},
		{"crypto/sha256", true},
		{"go/ast", true},
		{"archive/tar", true},
		{"compress/gzip", true},
		{"container/list", true},

		// 外部パッケージ
		{"github.com/user/repo", false},
		{"example.com/pkg", false},
		{"golang.org/x/tools", false},

		// 相対パス
		{"./relative", false},
		{"../relative", false},

		// エッジケース
		{"", false},
		{".", false},

		// 単一パッケージ名（標準ライブラリではない）
		{"mypackage", false},
		{"customlib", false},
	}

	for _, test := range tests {
		result := IsStandardLibrary(test.importPath)
		if result != test.expected {
			t.Errorf("IsStandardLibrary(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		importPath string
		expected   string
	}{
		{"github.com/user/repo/pkg", "pkg"},
		{"example.com/project/internal/service", "service"},
		{"pkg", "pkg"},
		{"./relative", "relative"},
		{"../parent", "parent"},
		{"fmt", "fmt"},
		{"net/http", "http"},
		{"encoding/json", "json"},
	}

	for _, test := range tests {
		result := ExtractPackageName(test.importPath)
		if result != test.expected {
			t.Errorf("ExtractPackageName(%s) = %s, expected %s", test.importPath, result, test.expected)
		}
	}
}

func TestIsLocalPackage(t *testing.T) {
	tests := []struct {
		importPath string
		expected   bool
	}{
		// 標準ライブラリ（ローカルではない）
		{"fmt", false},
		{"time", false},
		{"net/http", false},
		{"encoding/json", false},

		// 外部パッケージ（ローカル）
		{"github.com/user/repo", true},
		{"example.com/pkg", true},

		// 相対パス（ローカル）
		{"./relative", true},
		{"../relative", true},

		// エッジケース
		{"", true}, // 標準ライブラリではないのでローカル扱い
		{".", true},

		// 単一パッケージ名（標準ライブラリではないのでローカル扱い）
		{"mypackage", true},
		{"customlib", true},
	}

	for _, test := range tests {
		result := IsLocalPackage(test.importPath)
		if result != test.expected {
			t.Errorf("IsLocalPackage(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestGetStandardLibrariesFallback(t *testing.T) {
	fallbackLibs := getStandardLibrariesFallback()

	// 主要な標準ライブラリが含まれていることを確認
	expectedLibs := []string{
		"fmt", "os", "io", "time", "strings", "strconv",
		"context", "sync", "errors", "sort", "math",
		"net", "http", "encoding", "crypto",
	}

	for _, lib := range expectedLibs {
		if !fallbackLibs[lib] {
			t.Errorf("Expected standard library %s not found in fallback list", lib)
		}
	}

	// 非標準ライブラリが含まれていないことを確認
	nonStandardLibs := []string{
		"github.com/user/repo", "example.com/pkg", "mypackage",
	}

	for _, lib := range nonStandardLibs {
		if fallbackLibs[lib] {
			t.Errorf("Non-standard library %s found in fallback list", lib)
		}
	}
}
