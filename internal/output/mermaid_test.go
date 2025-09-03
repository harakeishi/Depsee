package output

import (
	"strings"
	"testing"

	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/types"
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
		NodeStabilities: map[types.NodeID]*graph.NodeStability{
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
	result := GenerateMermaid(g, stabilityResult)

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
	if !strings.Contains(result, `[📦 struct: graph<br>不安定度:1.00]`) {
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
		NodeStabilities: map[types.NodeID]*graph.NodeStability{
			"pkg.User-Service": {
				NodeID:      "pkg.User-Service",
				OutDegree:   0,
				InDegree:    0,
				Instability: 1.0,
			},
		},
	}

	result := GenerateMermaid(g, stabilityResult)

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

func TestGenerateMermaidWithPackageStability(t *testing.T) {
	// テスト用の依存グラフを作成（パッケージノードを含む）
	g := graph.NewDependencyGraph()

	// 通常のノードを追加
	userNode := &graph.Node{
		ID:      "pkg1.User",
		Kind:    graph.NodeStruct,
		Name:    "User",
		Package: "pkg1",
	}
	profileNode := &graph.Node{
		ID:      "pkg2.Profile",
		Kind:    graph.NodeStruct,
		Name:    "Profile",
		Package: "pkg2",
	}
	// パッケージノードを追加（除外されるべき）
	packageNode := &graph.Node{
		ID:      "package:pkg1",
		Kind:    graph.NodePackage,
		Name:    "pkg1",
		Package: "pkg1",
	}

	g.AddNode(userNode)
	g.AddNode(profileNode)
	g.AddNode(packageNode)

	// エッジを追加
	g.AddEdge("pkg1.User", "pkg2.Profile")    // 通常のノード間
	g.AddEdge("package:pkg1", "pkg2.Profile") // パッケージノードから（除外されるべき）

	// 安定度情報を作成
	stability := &graph.StabilityResult{
		NodeStabilities: map[types.NodeID]*graph.NodeStability{
			"pkg1.User": {
				NodeID:      "pkg1.User",
				OutDegree:   1,
				InDegree:    0,
				Instability: 1.0,
			},
			"pkg2.Profile": {
				NodeID:      "pkg2.Profile",
				OutDegree:   0,
				InDegree:    1,
				Instability: 0.0,
			},
			"package:pkg1": {
				NodeID:      "package:pkg1",
				OutDegree:   1,
				InDegree:    0,
				Instability: 1.0,
			},
		},
		PackageStabilities: map[string]*graph.PackageStability{
			"pkg1": {
				PackageName: "pkg1",
				OutDegree:   1,
				InDegree:    0,
				Instability: 1.0,
			},
			"pkg2": {
				PackageName: "pkg2",
				OutDegree:   0,
				InDegree:    1,
				Instability: 0.0,
			},
		},
	}

	// Mermaid出力を生成
	result := GenerateMermaid(g, stabilityResult)

	// 結果の検証
	if !strings.Contains(result, "graph TD") {
		t.Error("出力にgraph TDが含まれていません")
	}

	// パッケージの不安定度がサブグラフタイトルに含まれているかチェック
	if !strings.Contains(result, "pkg1 (不安定度:1.00)") {
		t.Error("pkg1の不安定度がサブグラフタイトルに表示されていません")
	}

	if !strings.Contains(result, "pkg2 (不安定度:0.00)") {
		t.Error("pkg2の不安定度がサブグラフタイトルに表示されていません")
	}

	// 通常のノードが含まれているかチェック
	if !strings.Contains(result, "pkg1_User") {
		t.Error("通常のノードpkg1.Userが含まれていません")
	}

	if !strings.Contains(result, "pkg2_Profile") {
		t.Error("通常のノードpkg2.Profileが含まれていません")
	}

	// パッケージノードが除外されているかチェック
	if strings.Contains(result, "package_pkg1") {
		t.Error("パッケージノードが除外されていません")
	}

	// 通常のノード間のエッジが含まれているかチェック
	if !strings.Contains(result, "pkg1_User --> pkg2_Profile") {
		t.Error("通常のノード間のエッジが含まれていません")
	}

	// パッケージノードからのエッジが除外されているかチェック
	if strings.Contains(result, "package_pkg1 -->") {
		t.Error("パッケージノードからのエッジが除外されていません")
	}

	t.Logf("生成されたMermaid出力:\n%s", result)
}

func TestGenerateMermaidWithSDPViolations(t *testing.T) {
	// SDP違反を含むテスト用の依存グラフを作成
	g := graph.NewDependencyGraph()

	// 不安定度が異なるノードを追加
	stableNode := &graph.Node{
		ID:      "pkg.Stable",
		Kind:    graph.NodeStruct,
		Name:    "Stable",
		Package: "pkg",
	}
	unstableNode := &graph.Node{
		ID:      "pkg.Unstable",
		Kind:    graph.NodeStruct,
		Name:    "Unstable",
		Package: "pkg",
	}
	normalNode := &graph.Node{
		ID:      "pkg.Normal",
		Kind:    graph.NodeStruct,
		Name:    "Normal",
		Package: "pkg",
	}

	g.AddNode(stableNode)
	g.AddNode(unstableNode)
	g.AddNode(normalNode)

	// エッジを追加（SDP違反を含む）
	g.AddEdge("pkg.Stable", "pkg.Unstable") // SDP違反: 安定 → 不安定
	g.AddEdge("pkg.Unstable", "pkg.Normal") // 正常: 不安定 → 中間
	g.AddEdge("pkg.Normal", "pkg.Stable")   // 正常: 中間 → 安定

	// 安定度情報を作成（SDP違反が発生するように設定）
	stability := &graph.StabilityResult{
		NodeStabilities: map[types.NodeID]*graph.NodeStability{
			"pkg.Stable": {
				NodeID:      "pkg.Stable",
				OutDegree:   1,   // 1つに依存
				InDegree:    1,   // 1つから依存される
				Instability: 0.5, // 中間の不安定度
			},
			"pkg.Unstable": {
				NodeID:      "pkg.Unstable",
				OutDegree:   1,   // 1つに依存
				InDegree:    1,   // 1つから依存される
				Instability: 0.5, // 中間の不安定度（実際は同じだが、テストのため）
			},
			"pkg.Normal": {
				NodeID:      "pkg.Normal",
				OutDegree:   1,   // 1つに依存
				InDegree:    1,   // 1つから依存される
				Instability: 0.5, // 中間の不安定度
			},
		},
		SDPViolations: []graph.SDPViolation{
			{
				From:              "pkg.Stable",
				To:                "pkg.Unstable",
				FromInstability:   0.2, // より安定
				ToInstability:     0.8, // より不安定
				ViolationSeverity: 0.6,
			},
		},
	}

	// SDP違反ハイライトなしでMermaid出力を生成
	resultWithoutHighlight := GenerateMermaidWithOptions(g, stabilityResult, false)

	// SDP違反ハイライトありでMermaid出力を生成
	resultWithHighlight := GenerateMermaidWithOptions(g, stabilityResult, true)

	// 結果の検証
	if !strings.Contains(resultWithoutHighlight, "graph TD") {
		t.Error("ハイライトなしの出力にgraph TDが含まれていません")
	}

	if !strings.Contains(resultWithHighlight, "graph TD") {
		t.Error("ハイライトありの出力にgraph TDが含まれていません")
	}

	// SDP違反ハイライトなしの場合、linkStyleが含まれていないことを確認
	if strings.Contains(resultWithoutHighlight, "linkStyle") {
		t.Error("ハイライトなしの出力にlinkStyleが含まれています")
	}

	// SDP違反ハイライトありの場合、linkStyleが含まれていることを確認
	if !strings.Contains(resultWithHighlight, "linkStyle") {
		t.Error("ハイライトありの出力にlinkStyleが含まれていません")
	}

	// 赤色のスタイルが適用されていることを確認
	if !strings.Contains(resultWithHighlight, "stroke:#ff0000") {
		t.Error("SDP違反エッジに赤色のスタイルが適用されていません")
	}

	t.Logf("SDP違反ハイライトなしの出力:\n%s", resultWithoutHighlight)
	t.Logf("SDP違反ハイライトありの出力:\n%s", resultWithHighlight)
}
