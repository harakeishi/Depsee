package graph

import (
	"math"
	"testing"
)

func TestCalculateStability(t *testing.T) {
	g := NewDependencyGraph()

	// テスト用のノードを作成
	nodes := []*Node{
		{ID: "test.A", Kind: NodeStruct, Name: "A", Package: "test"},
		{ID: "test.B", Kind: NodeStruct, Name: "B", Package: "test"},
		{ID: "test.C", Kind: NodeStruct, Name: "C", Package: "test"},
		{ID: "test.D", Kind: NodeStruct, Name: "D", Package: "test"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// 依存関係を追加
	// A -> B, A -> C (Aは2つに依存)
	// B -> C (Bは1つに依存)
	// D -> A (DはAに依存)
	g.AddEdge("test.A", "test.B")
	g.AddEdge("test.A", "test.C")
	g.AddEdge("test.B", "test.C")
	g.AddEdge("test.D", "test.A")

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	if len(result.NodeStabilities) != 4 {
		t.Errorf("Expected 4 node stabilities, got %d", len(result.NodeStabilities))
	}

	// ノードAの安定度検証
	// A: Ce=2 (A->B, A->C), Ca=1 (D->A), I=Ce/(Ce+Ca)=2/3≈0.67
	aStability := result.NodeStabilities["test.A"]
	if aStability == nil {
		t.Fatal("Node A stability not found")
	}

	if aStability.OutDegree != 2 {
		t.Errorf("Node A OutDegree: expected 2, got %d", aStability.OutDegree)
	}

	if aStability.InDegree != 1 {
		t.Errorf("Node A InDegree: expected 1, got %d", aStability.InDegree)
	}

	expectedInstability := 2.0 / 3.0
	if math.Abs(aStability.Instability-expectedInstability) > 0.001 {
		t.Errorf("Node A Instability: expected %.3f, got %.3f",
			expectedInstability, aStability.Instability)
	}

	// ノードCの安定度検証
	// C: Ce=0, Ca=2 (A->C, B->C), I=0/(0+2)=0 (最も安定)
	cStability := result.NodeStabilities["test.C"]
	if cStability == nil {
		t.Fatal("Node C stability not found")
	}

	if cStability.OutDegree != 0 {
		t.Errorf("Node C OutDegree: expected 0, got %d", cStability.OutDegree)
	}

	if cStability.InDegree != 2 {
		t.Errorf("Node C InDegree: expected 2, got %d", cStability.InDegree)
	}

	if cStability.Instability != 0.0 {
		t.Errorf("Node C Instability: expected 0.0, got %.3f", cStability.Instability)
	}

	// ノードDの安定度検証
	// D: Ce=1 (D->A), Ca=0, I=1/(1+0)=1 (最も不安定)
	dStability := result.NodeStabilities["test.D"]
	if dStability == nil {
		t.Fatal("Node D stability not found")
	}

	if dStability.OutDegree != 1 {
		t.Errorf("Node D OutDegree: expected 1, got %d", dStability.OutDegree)
	}

	if dStability.InDegree != 0 {
		t.Errorf("Node D InDegree: expected 0, got %d", dStability.InDegree)
	}

	if dStability.Instability != 1.0 {
		t.Errorf("Node D Instability: expected 1.0, got %.3f", dStability.Instability)
	}
}

func TestCalculateStabilityIsolatedNode(t *testing.T) {
	g := NewDependencyGraph()

	// 孤立したノードを追加
	node := &Node{ID: "test.Isolated", Kind: NodeStruct, Name: "Isolated", Package: "test"}
	g.AddNode(node)

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	stability := result.NodeStabilities["test.Isolated"]
	if stability == nil {
		t.Fatal("Isolated node stability not found")
	}

	// 孤立ノードは不安定度1.0とする
	if stability.Instability != 1.0 {
		t.Errorf("Isolated node Instability: expected 1.0, got %.3f", stability.Instability)
	}

	if stability.OutDegree != 0 {
		t.Errorf("Isolated node OutDegree: expected 0, got %d", stability.OutDegree)
	}

	if stability.InDegree != 0 {
		t.Errorf("Isolated node InDegree: expected 0, got %d", stability.InDegree)
	}
}

func TestCalculateStabilityEmptyGraph(t *testing.T) {
	g := NewDependencyGraph()

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	if len(result.NodeStabilities) != 0 {
		t.Errorf("Expected 0 node stabilities for empty graph, got %d",
			len(result.NodeStabilities))
	}
}

func TestCalculateStabilityLinearChain(t *testing.T) {
	g := NewDependencyGraph()

	// 線形チェーンを作成: A -> B -> C
	nodes := []*Node{
		{ID: "test.A", Kind: NodeStruct, Name: "A", Package: "test"},
		{ID: "test.B", Kind: NodeStruct, Name: "B", Package: "test"},
		{ID: "test.C", Kind: NodeStruct, Name: "C", Package: "test"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	g.AddEdge("test.A", "test.B")
	g.AddEdge("test.B", "test.C")

	result := CalculateStability(g)

	// A: Ce=1, Ca=0, I=1.0 (最も不安定)
	aStability := result.NodeStabilities["test.A"]
	if aStability.Instability != 1.0 {
		t.Errorf("Node A Instability: expected 1.0, got %.3f", aStability.Instability)
	}

	// B: Ce=1, Ca=1, I=0.5 (中間)
	bStability := result.NodeStabilities["test.B"]
	if bStability.Instability != 0.5 {
		t.Errorf("Node B Instability: expected 0.5, got %.3f", bStability.Instability)
	}

	// C: Ce=0, Ca=1, I=0.0 (最も安定)
	cStability := result.NodeStabilities["test.C"]
	if cStability.Instability != 0.0 {
		t.Errorf("Node C Instability: expected 0.0, got %.3f", cStability.Instability)
	}
}

func TestNodeStabilityStruct(t *testing.T) {
	stability := &NodeStability{
		NodeID:      "test.Node",
		OutDegree:   3,
		InDegree:    2,
		Instability: 0.6,
	}

	if stability.NodeID != "test.Node" {
		t.Errorf("NodeID: expected 'test.Node', got '%s'", stability.NodeID)
	}

	if stability.OutDegree != 3 {
		t.Errorf("OutDegree: expected 3, got %d", stability.OutDegree)
	}

	if stability.InDegree != 2 {
		t.Errorf("InDegree: expected 2, got %d", stability.InDegree)
	}

	if stability.Instability != 0.6 {
		t.Errorf("Instability: expected 0.6, got %.3f", stability.Instability)
	}
}

func TestCalculatePackageStability(t *testing.T) {
	g := NewDependencyGraph()

	// 複数パッケージのテスト用ノードを作成
	nodes := []*Node{
		// pkg1のノード
		{ID: "pkg1.A", Kind: NodeStruct, Name: "A", Package: "pkg1"},
		{ID: "pkg1.B", Kind: NodeInterface, Name: "B", Package: "pkg1"},
		// pkg2のノード
		{ID: "pkg2.C", Kind: NodeStruct, Name: "C", Package: "pkg2"},
		{ID: "pkg2.D", Kind: NodeFunc, Name: "D", Package: "pkg2"},
		// pkg3のノード
		{ID: "pkg3.E", Kind: NodeStruct, Name: "E", Package: "pkg3"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// パッケージ間依存関係を作成
	// pkg1.A -> pkg2.C (pkg1 -> pkg2)
	// pkg1.B -> pkg2.D (pkg1 -> pkg2)
	// pkg2.C -> pkg3.E (pkg2 -> pkg3)
	// pkg1内の依存関係: pkg1.A -> pkg1.B (パッケージ間依存には含まれない)
	g.AddEdge("pkg1.A", "pkg2.C")
	g.AddEdge("pkg1.B", "pkg2.D")
	g.AddEdge("pkg2.C", "pkg3.E")
	g.AddEdge("pkg1.A", "pkg1.B") // 同一パッケージ内

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	if len(result.PackageStabilities) != 3 {
		t.Errorf("Expected 3 package stabilities, got %d", len(result.PackageStabilities))
	}

	// pkg1の安定度検証
	// pkg1: Ce=1 (pkg1->pkg2), Ca=0, I=1.0 (最も不安定)
	pkg1Stability := result.PackageStabilities["pkg1"]
	if pkg1Stability == nil {
		t.Fatal("Package pkg1 stability not found")
	}

	if pkg1Stability.OutDegree != 1 {
		t.Errorf("Package pkg1 OutDegree: expected 1, got %d", pkg1Stability.OutDegree)
	}

	if pkg1Stability.InDegree != 0 {
		t.Errorf("Package pkg1 InDegree: expected 0, got %d", pkg1Stability.InDegree)
	}

	if pkg1Stability.Instability != 1.0 {
		t.Errorf("Package pkg1 Instability: expected 1.0, got %.3f", pkg1Stability.Instability)
	}

	// pkg2の安定度検証
	// pkg2: Ce=1 (pkg2->pkg3), Ca=1 (pkg1->pkg2), I=0.5 (中間)
	pkg2Stability := result.PackageStabilities["pkg2"]
	if pkg2Stability == nil {
		t.Fatal("Package pkg2 stability not found")
	}

	if pkg2Stability.OutDegree != 1 {
		t.Errorf("Package pkg2 OutDegree: expected 1, got %d", pkg2Stability.OutDegree)
	}

	if pkg2Stability.InDegree != 1 {
		t.Errorf("Package pkg2 InDegree: expected 1, got %d", pkg2Stability.InDegree)
	}

	if pkg2Stability.Instability != 0.5 {
		t.Errorf("Package pkg2 Instability: expected 0.5, got %.3f", pkg2Stability.Instability)
	}

	// pkg3の安定度検証
	// pkg3: Ce=0, Ca=1 (pkg2->pkg3), I=0.0 (最も安定)
	pkg3Stability := result.PackageStabilities["pkg3"]
	if pkg3Stability == nil {
		t.Fatal("Package pkg3 stability not found")
	}

	if pkg3Stability.OutDegree != 0 {
		t.Errorf("Package pkg3 OutDegree: expected 0, got %d", pkg3Stability.OutDegree)
	}

	if pkg3Stability.InDegree != 1 {
		t.Errorf("Package pkg3 InDegree: expected 1, got %d", pkg3Stability.InDegree)
	}

	if pkg3Stability.Instability != 0.0 {
		t.Errorf("Package pkg3 Instability: expected 0.0, got %.3f", pkg3Stability.Instability)
	}
}

func TestCalculatePackageStabilityWithPackageNodes(t *testing.T) {
	g := NewDependencyGraph()

	// パッケージノードと通常のノードを混在させたテスト
	nodes := []*Node{
		// 通常のノード
		{ID: "pkg1.A", Kind: NodeStruct, Name: "A", Package: "pkg1"},
		{ID: "pkg2.B", Kind: NodeStruct, Name: "B", Package: "pkg2"},
		// パッケージノード（除外されるべき）
		{ID: "package:pkg1", Kind: NodePackage, Name: "pkg1", Package: "pkg1"},
		{ID: "package:pkg2", Kind: NodePackage, Name: "pkg2", Package: "pkg2"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// 依存関係を追加（パッケージノード間の依存も含む）
	g.AddEdge("pkg1.A", "pkg2.B")             // 通常のノード間
	g.AddEdge("package:pkg1", "package:pkg2") // パッケージノード間（除外されるべき）

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	// パッケージノード間の依存関係は除外され、通常のノード間の依存関係のみが考慮される
	if len(result.PackageStabilities) != 2 {
		t.Errorf("Expected 2 package stabilities, got %d", len(result.PackageStabilities))
	}

	// pkg1の安定度検証（パッケージノードの依存関係は除外される）
	pkg1Stability := result.PackageStabilities["pkg1"]
	if pkg1Stability == nil {
		t.Fatal("Package pkg1 stability not found")
	}

	if pkg1Stability.OutDegree != 1 {
		t.Errorf("Package pkg1 OutDegree: expected 1, got %d", pkg1Stability.OutDegree)
	}

	if pkg1Stability.InDegree != 0 {
		t.Errorf("Package pkg1 InDegree: expected 0, got %d", pkg1Stability.InDegree)
	}

	if pkg1Stability.Instability != 1.0 {
		t.Errorf("Package pkg1 Instability: expected 1.0, got %.3f", pkg1Stability.Instability)
	}
}

func TestCalculatePackageStabilityWithDirectPackageDependencies(t *testing.T) {
	g := NewDependencyGraph()

	// 通常のノードとパッケージノードを混在させたテスト
	nodes := []*Node{
		// 通常のノード
		{ID: "pkg1.A", Kind: NodeStruct, Name: "A", Package: "pkg1"},
		{ID: "pkg2.B", Kind: NodeStruct, Name: "B", Package: "pkg2"},
		{ID: "pkg3.C", Kind: NodeStruct, Name: "C", Package: "pkg3"},
		// パッケージノード
		{ID: "package:pkg1", Kind: NodePackage, Name: "pkg1", Package: "pkg1"},
		{ID: "package:pkg2", Kind: NodePackage, Name: "pkg2", Package: "pkg2"},
		{ID: "package:pkg3", Kind: NodePackage, Name: "pkg3", Package: "pkg3"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// 依存関係を追加
	// 通常のノード間: pkg1.A -> pkg2.B (pkg1 -> pkg2)
	// パッケージノード間: package:pkg2 -> package:pkg3 (pkg2 -> pkg3)
	g.AddEdge("pkg1.A", "pkg2.B")
	g.AddEdge("package:pkg2", "package:pkg3")

	result := CalculateStability(g)

	if result == nil {
		t.Fatal("CalculateStability returned nil")
	}

	if len(result.PackageStabilities) != 3 {
		t.Errorf("Expected 3 package stabilities, got %d", len(result.PackageStabilities))
	}

	// pkg1の安定度検証
	// pkg1: Ce=1 (pkg1->pkg2), Ca=0, I=1.0 (最も不安定)
	pkg1Stability := result.PackageStabilities["pkg1"]
	if pkg1Stability == nil {
		t.Fatal("Package pkg1 stability not found")
	}

	if pkg1Stability.OutDegree != 1 {
		t.Errorf("Package pkg1 OutDegree: expected 1, got %d", pkg1Stability.OutDegree)
	}

	if pkg1Stability.InDegree != 0 {
		t.Errorf("Package pkg1 InDegree: expected 0, got %d", pkg1Stability.InDegree)
	}

	if pkg1Stability.Instability != 1.0 {
		t.Errorf("Package pkg1 Instability: expected 1.0, got %.3f", pkg1Stability.Instability)
	}

	// pkg2の安定度検証
	// pkg2: Ce=1 (pkg2->pkg3), Ca=1 (pkg1->pkg2), I=0.5 (中間)
	pkg2Stability := result.PackageStabilities["pkg2"]
	if pkg2Stability == nil {
		t.Fatal("Package pkg2 stability not found")
	}

	if pkg2Stability.OutDegree != 1 {
		t.Errorf("Package pkg2 OutDegree: expected 1, got %d", pkg2Stability.OutDegree)
	}

	if pkg2Stability.InDegree != 1 {
		t.Errorf("Package pkg2 InDegree: expected 1, got %d", pkg2Stability.InDegree)
	}

	if pkg2Stability.Instability != 0.5 {
		t.Errorf("Package pkg2 Instability: expected 0.5, got %.3f", pkg2Stability.Instability)
	}

	// pkg3の安定度検証
	// pkg3: Ce=0, Ca=1 (pkg2->pkg3), I=0.0 (最も安定)
	pkg3Stability := result.PackageStabilities["pkg3"]
	if pkg3Stability == nil {
		t.Fatal("Package pkg3 stability not found")
	}

	if pkg3Stability.OutDegree != 0 {
		t.Errorf("Package pkg3 OutDegree: expected 0, got %d", pkg3Stability.OutDegree)
	}

	if pkg3Stability.InDegree != 1 {
		t.Errorf("Package pkg3 InDegree: expected 1, got %d", pkg3Stability.InDegree)
	}

	if pkg3Stability.Instability != 0.0 {
		t.Errorf("Package pkg3 Instability: expected 0.0, got %.3f", pkg3Stability.Instability)
	}
}

func TestExtractPackageNameFromNodeID(t *testing.T) {
	tests := []struct {
		name     string
		nodeID   string
		expected string
	}{
		{
			name:     "パッケージノードID",
			nodeID:   "package:cmd",
			expected: "cmd",
		},
		{
			name:     "パッケージノードID（複雑なパッケージ名）",
			nodeID:   "package:github.com/example/pkg",
			expected: "github.com/example/pkg",
		},
		{
			name:     "通常のノードID",
			nodeID:   "cmd.Execute",
			expected: "",
		},
		{
			name:     "空文字列",
			nodeID:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPackageNameFromNodeID(tt.nodeID)
			if result != tt.expected {
				t.Errorf("extractPackageNameFromNodeID(%q) = %q, want %q", tt.nodeID, result, tt.expected)
			}
		})
	}
}

func TestDetectSDPViolations(t *testing.T) {
	g := NewDependencyGraph()

	// SDP違反を含むテスト用のノードを作成
	nodes := []*Node{
		{ID: "test.Stable", Kind: NodeStruct, Name: "Stable", Package: "test"},
		{ID: "test.Unstable", Kind: NodeStruct, Name: "Unstable", Package: "test"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// 依存関係を追加（SDP違反を含む）
	// Stable -> Unstable (SDP違反: 安定 -> 不安定)
	// Stable: Ce=1, Ca=0, I=1.0 (不安定)
	// Unstable: Ce=0, Ca=1, I=0.0 (安定)
	// これは逆転しているため、実際にはStableが不安定でUnstableが安定になる
	// 正しいSDP違反を作るため、逆にする
	g.AddEdge("test.Unstable", "test.Stable") // 実際は Unstable(不安定) -> Stable(安定) で正常

	// 実際のSDP違反を作るため、追加のノードと依存関係を作成
	additionalNode := &Node{ID: "test.VeryUnstable", Kind: NodeStruct, Name: "VeryUnstable", Package: "test"}
	g.AddNode(additionalNode)

	// VeryUnstable -> Stable (正常: 不安定 -> 安定)
	// Stable -> VeryUnstable (SDP違反: 安定 -> 不安定)
	g.AddEdge("test.VeryUnstable", "test.Stable")
	g.AddEdge("test.Stable", "test.VeryUnstable") // これがSDP違反になる

	result := CalculateStability(g)

	// 不安定度を確認
	t.Logf("Stable不安定度: %.3f", result.NodeStabilities["test.Stable"].Instability)
	t.Logf("Unstable不安定度: %.3f", result.NodeStabilities["test.Unstable"].Instability)
	t.Logf("VeryUnstable不安定度: %.3f", result.NodeStabilities["test.VeryUnstable"].Instability)

	// SDP違反が検出されることを確認
	if len(result.SDPViolations) == 0 {
		t.Error("SDP違反が検出されませんでした")
	}

	t.Logf("検出されたSDP違反数: %d", len(result.SDPViolations))
	for i, violation := range result.SDPViolations {
		t.Logf("違反%d: %s (%.3f) -> %s (%.3f), 違反度: %.3f",
			i+1, violation.From, violation.FromInstability,
			violation.To, violation.ToInstability, violation.ViolationSeverity)
	}

	// 少なくとも1つのSDP違反があることを確認
	if len(result.SDPViolations) == 0 {
		t.Error("SDP違反が検出されませんでした")
	}

	// 各違反について、不安定度の関係が正しいことを確認
	for _, violation := range result.SDPViolations {
		if violation.FromInstability >= violation.ToInstability {
			t.Errorf("SDP違反の不安定度関係が正しくありません: From=%.3f, To=%.3f",
				violation.FromInstability, violation.ToInstability)
		}

		// 違反度が正しく計算されているか確認
		expectedSeverity := violation.ToInstability - violation.FromInstability
		if math.Abs(violation.ViolationSeverity-expectedSeverity) > 0.001 {
			t.Errorf("違反度が正しく計算されていません: expected=%.3f, got=%.3f",
				expectedSeverity, violation.ViolationSeverity)
		}
	}
}

func TestDetectSDPViolationsNoViolations(t *testing.T) {
	g := NewDependencyGraph()

	// SDP違反のない理想的な依存関係を作成
	nodes := []*Node{
		{ID: "test.A", Kind: NodeStruct, Name: "A", Package: "test"},
		{ID: "test.B", Kind: NodeStruct, Name: "B", Package: "test"},
		{ID: "test.C", Kind: NodeStruct, Name: "C", Package: "test"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// 理想的な依存関係: 不安定 -> 安定
	// A -> B -> C (Aが最も不安定、Cが最も安定)
	g.AddEdge("test.A", "test.B")
	g.AddEdge("test.B", "test.C")

	result := CalculateStability(g)

	// SDP違反がないことを確認
	if len(result.SDPViolations) != 0 {
		t.Errorf("SDP違反が検出されましたが、期待されていません。違反数: %d", len(result.SDPViolations))
		for i, violation := range result.SDPViolations {
			t.Logf("予期しない違反%d: %s (%.3f) -> %s (%.3f)",
				i+1, violation.From, violation.FromInstability,
				violation.To, violation.ToInstability)
		}
	}
}

func TestDetectSDPViolationsEmptyGraph(t *testing.T) {
	g := NewDependencyGraph()

	result := CalculateStability(g)

	// 空のグラフではSDP違反はない
	if len(result.SDPViolations) != 0 {
		t.Errorf("空のグラフでSDP違反が検出されました。違反数: %d", len(result.SDPViolations))
	}
}
