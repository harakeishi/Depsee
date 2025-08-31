package graph

import (
	"os"
	"testing"

	"github.com/harakeishi/depsee/internal/analyzer"
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

func TestNewDependencyGraph(t *testing.T) {
	g := NewDependencyGraph()

	if g == nil {
		t.Fatal("NewDependencyGraph returned nil")
	}

	if g.Nodes == nil {
		t.Error("Nodes map is nil")
	}

	if g.Edges == nil {
		t.Error("Edges map is nil")
	}

	if len(g.Nodes) != 0 {
		t.Errorf("Expected empty nodes, got %d", len(g.Nodes))
	}

	if len(g.Edges) != 0 {
		t.Errorf("Expected empty edges, got %d", len(g.Edges))
	}
}

func TestAddNode(t *testing.T) {
	g := NewDependencyGraph()

	node := &Node{
		ID:      "test.TestStruct",
		Kind:    NodeStruct,
		Name:    "TestStruct",
		Package: "test",
	}

	g.AddNode(node)

	if len(g.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(g.Nodes))
	}

	retrievedNode, exists := g.Nodes["test.TestStruct"]
	if !exists {
		t.Error("Node was not added to the graph")
	}

	if retrievedNode.Name != "TestStruct" {
		t.Errorf("Expected node name 'TestStruct', got '%s'", retrievedNode.Name)
	}
}

func TestAddEdge(t *testing.T) {
	g := NewDependencyGraph()

	// ノードを追加
	node1 := &Node{ID: "test.A", Kind: NodeStruct, Name: "A", Package: "test"}
	node2 := &Node{ID: "test.B", Kind: NodeStruct, Name: "B", Package: "test"}
	g.AddNode(node1)
	g.AddNode(node2)

	// エッジを追加
	g.AddEdge("test.A", "test.B")

	if len(g.Edges) != 1 {
		t.Errorf("Expected 1 edge source, got %d", len(g.Edges))
	}

	targets, exists := g.Edges["test.A"]
	if !exists {
		t.Error("Edge was not added to the graph")
	}

	if len(targets) != 1 {
		t.Errorf("Expected 1 edge target, got %d", len(targets))
	}

	_, targetExists := targets["test.B"]
	if !targetExists {
		t.Error("Expected target 'test.B' not found")
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	// テスト用の解析結果を作成
	result := &analyzer.Result{
		Structs: []analyzer.StructInfo{
			{
				Name:    "User",
				Package: "test",
				File:    "test.go",
				Fields: []analyzer.FieldInfo{
					{Name: "ID", Type: "int"},
					{Name: "Profile", Type: "*Profile"},
				},
			},
			{
				Name:    "Profile",
				Package: "test",
				File:    "test.go",
				Fields: []analyzer.FieldInfo{
					{Name: "Bio", Type: "string"},
				},
			},
		},
		Interfaces: []analyzer.InterfaceInfo{
			{
				Name:    "UserService",
				Package: "test",
				File:    "test.go",
			},
		},
		Functions: []analyzer.FuncInfo{
			{
				Name:    "CreateUser",
				Package: "test",
				File:    "test.go",
				Params: []analyzer.FieldInfo{
					{Name: "name", Type: "string"},
				},
				Results: []analyzer.FieldInfo{
					{Name: "", Type: "*User"},
				},
			},
		},
	}

	g := BuildDependencyGraph(result)

	if g == nil {
		t.Fatal("BuildDependencyGraph returned nil")
	}

	// ノード数の検証
	expectedNodeCount := 3 // User, Profile, UserService, CreateUser
	if len(g.Nodes) < expectedNodeCount {
		t.Errorf("Expected at least %d nodes, got %d", expectedNodeCount, len(g.Nodes))
	}

	// 特定のノードの存在確認
	expectedNodes := []string{"test.User", "test.Profile", "test.UserService", "test.CreateUser"}
	for _, nodeID := range expectedNodes {
		if _, exists := g.Nodes[NodeID(nodeID)]; !exists {
			t.Errorf("Expected node '%s' not found", nodeID)
		}
	}

	// 依存関係の検証（User -> Profile）
	userEdges, exists := g.Edges["test.User"]
	if !exists {
		t.Error("User node should have outgoing edges")
	} else {
		if _, profileDep := userEdges["test.Profile"]; !profileDep {
			t.Error("User should depend on Profile")
		}
	}
}

func TestRegisterNodes(t *testing.T) {
	g := NewDependencyGraph()

	result := &analyzer.Result{
		Structs: []analyzer.StructInfo{
			{Name: "TestStruct", Package: "test", File: "test.go"},
		},
		Interfaces: []analyzer.InterfaceInfo{
			{Name: "TestInterface", Package: "test", File: "test.go"},
		},
		Functions: []analyzer.FuncInfo{
			{Name: "TestFunc", Package: "test", File: "test.go"},
		},
	}

	registerNodes(result, g)

	expectedNodes := map[string]NodeKind{
		"test.TestStruct":    NodeStruct,
		"test.TestInterface": NodeInterface,
		"test.TestFunc":      NodeFunc,
	}

	if len(g.Nodes) != len(expectedNodes) {
		t.Errorf("Expected %d nodes, got %d", len(expectedNodes), len(g.Nodes))
	}

	for nodeID, expectedKind := range expectedNodes {
		node, exists := g.Nodes[NodeID(nodeID)]
		if !exists {
			t.Errorf("Expected node '%s' not found", nodeID)
			continue
		}

		if node.Kind != expectedKind {
			t.Errorf("Node '%s' has wrong kind. Expected %d, got %d",
				nodeID, expectedKind, node.Kind)
		}
	}
}

func TestCountEdges(t *testing.T) {
	g := NewDependencyGraph()

	// ノードを追加
	nodes := []*Node{
		{ID: "test.A", Kind: NodeStruct, Name: "A", Package: "test"},
		{ID: "test.B", Kind: NodeStruct, Name: "B", Package: "test"},
		{ID: "test.C", Kind: NodeStruct, Name: "C", Package: "test"},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// エッジを追加
	g.AddEdge("test.A", "test.B")
	g.AddEdge("test.A", "test.C")
	g.AddEdge("test.B", "test.C")

	edgeCount := countEdges(g)
	expectedCount := 3

	if edgeCount != expectedCount {
		t.Errorf("Expected %d edges, got %d", expectedCount, edgeCount)
	}
}
