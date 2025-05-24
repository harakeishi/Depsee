package output

import (
	"strings"
	"testing"

	"github.com/harakeishi/depsee/internal/graph"
)

func TestSanitizeNodeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のID",
			input:    "User",
			expected: "User",
		},
		{
			name:     "パッケージ名付きID",
			input:    "sample.User",
			expected: "sample_User",
		},
		{
			name:     "予約語graph",
			input:    "graph",
			expected: "node_graph",
		},
		{
			name:     "予約語end",
			input:    "end",
			expected: "node_end",
		},
		{
			name:     "予約語default",
			input:    "default",
			expected: "node_default",
		},
		{
			name:     "大文字の予約語",
			input:    "GRAPH",
			expected: "node_GRAPH",
		},
		{
			name:     "数字で始まるID",
			input:    "123User",
			expected: "node_123User",
		},
		{
			name:     "特殊文字を含むID",
			input:    "User-Service",
			expected: "User_Service",
		},
		{
			name:     "複雑な特殊文字",
			input:    "User@Service#1",
			expected: "User_Service_1",
		},
		{
			name:     "パッケージ名と予約語の組み合わせ",
			input:    "pkg.graph",
			expected: "node_pkg_graph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeNodeID(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeNodeID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeNodeLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のラベル",
			input:    "User",
			expected: "User",
		},
		{
			name:     "HTMLタグを含むラベル",
			input:    "<User>",
			expected: "&lt;User&gt;",
		},
		{
			name:     "引用符を含むラベル",
			input:    `User "Service"`,
			expected: "User &quot;Service&quot;",
		},
		{
			name:     "アンパサンドを含むラベル",
			input:    "User & Service",
			expected: "User &amp; Service",
		},
		{
			name:     "シングルクォートを含むラベル",
			input:    "User's Service",
			expected: "User&#39;s Service",
		},
		{
			name:     "複数の特殊文字",
			input:    `<User & "Service">`,
			expected: "&lt;User &amp; &quot;Service&quot;&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeNodeLabel(tt.input)
			if result != tt.expected {
				t.Errorf("escapeNodeLabel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateMermaidWithReservedWords(t *testing.T) {
	// テスト用の依存グラフを作成（予約語を含む）
	g := graph.NewDependencyGraph()

	// 予約語を含むノードを追加
	graphNode := &graph.Node{
		ID:      "pkg.graph",
		Kind:    graph.NodeStruct,
		Name:    "graph",
		Package: "pkg",
	}
	endNode := &graph.Node{
		ID:      "pkg.end",
		Kind:    graph.NodeStruct,
		Name:    "end",
		Package: "pkg",
	}
	userNode := &graph.Node{
		ID:      "pkg.User",
		Kind:    graph.NodeStruct,
		Name:    "User",
		Package: "pkg",
	}

	g.AddNode(graphNode)
	g.AddNode(endNode)
	g.AddNode(userNode)

	// エッジを追加
	g.AddEdge("pkg.graph", "pkg.User")
	g.AddEdge("pkg.end", "pkg.User")

	// 安定度情報を作成
	stability := &graph.StabilityResult{
		NodeStabilities: map[graph.NodeID]*graph.NodeStability{
			"pkg.graph": {
				NodeID:      "pkg.graph",
				OutDegree:   1,
				InDegree:    0,
				Instability: 1.0,
			},
			"pkg.end": {
				NodeID:      "pkg.end",
				OutDegree:   1,
				InDegree:    0,
				Instability: 1.0,
			},
			"pkg.User": {
				NodeID:      "pkg.User",
				OutDegree:   0,
				InDegree:    2,
				Instability: 0.0,
			},
		},
	}

	// Mermaid出力を生成
	result := GenerateMermaid(g, stability)

	// 結果の検証
	if !strings.Contains(result, "graph TD") {
		t.Error("出力にgraph TDが含まれていません")
	}

	// 予約語がエスケープされているかチェック
	if !strings.Contains(result, "node_pkg_graph") {
		t.Error("予約語graphがエスケープされていません")
	}

	if !strings.Contains(result, "node_pkg_end") {
		t.Error("予約語endがエスケープされていません")
	}

	// 通常のノードIDは変更されていないかチェック
	if !strings.Contains(result, "pkg_User") {
		t.Error("通常のノードIDが正しく処理されていません")
	}

	// エッジが正しく出力されているかチェック
	if !strings.Contains(result, "node_pkg_graph --> pkg_User") {
		t.Error("予約語を含むエッジが正しく出力されていません")
	}

	if !strings.Contains(result, "node_pkg_end --> pkg_User") {
		t.Error("予約語を含むエッジが正しく出力されていません")
	}

	// ラベルが正しくエスケープされているかチェック
	if !strings.Contains(result, `["graph<br>不安定度:1.00"]`) {
		t.Error("ノードラベルが正しく出力されていません")
	}

	t.Logf("生成されたMermaid出力:\n%s", result)
}

func TestGenerateMermaidWithSpecialCharacters(t *testing.T) {
	// 特殊文字を含むテストケース
	g := graph.NewDependencyGraph()

	// 特殊文字を含むノードを追加
	specialNode := &graph.Node{
		ID:      "pkg.User-Service",
		Kind:    graph.NodeStruct,
		Name:    `User "Service" & <Component>`,
		Package: "pkg",
	}

	g.AddNode(specialNode)

	stability := &graph.StabilityResult{
		NodeStabilities: map[graph.NodeID]*graph.NodeStability{
			"pkg.User-Service": {
				NodeID:      "pkg.User-Service",
				OutDegree:   0,
				InDegree:    0,
				Instability: 1.0,
			},
		},
	}

	result := GenerateMermaid(g, stability)

	// 特殊文字がエスケープされているかチェック
	if !strings.Contains(result, "pkg_User_Service") {
		t.Error("特殊文字を含むノードIDが正しく処理されていません")
	}

	// ラベルの特殊文字がエスケープされているかチェック
	if !strings.Contains(result, "&quot;") && !strings.Contains(result, "&amp;") && !strings.Contains(result, "&lt;") {
		t.Error("ラベルの特殊文字が正しくエスケープされていません")
	}

	t.Logf("特殊文字テストの出力:\n%s", result)
}
