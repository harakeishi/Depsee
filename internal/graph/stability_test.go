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
