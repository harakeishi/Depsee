package output

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/harakeishi/depsee/internal/graph"
)

// mermaidの予約語リスト
var mermaidReservedWords = map[string]bool{
	// フローチャート関連
	"graph":     true,
	"flowchart": true,
	"subgraph":  true,
	"end":       true,
	"default":   true,

	// ノード形状
	"circle":   true,
	"rect":     true,
	"diamond":  true,
	"hexagon":  true,
	"stadium":  true,
	"cylinder": true,

	// 方向
	"TD": true,
	"TB": true,
	"BT": true,
	"LR": true,
	"RL": true,

	// その他の予約語
	"class":     true,
	"classDef":  true,
	"click":     true,
	"style":     true,
	"linkStyle": true,
	"fill":      true,
	"stroke":    true,
	"color":     true,
	"node":      true,
	"edge":      true,
	"link":      true,
}

// sanitizeNodeID はノードIDをmermaidで安全に使用できるように変換する
func sanitizeNodeID(id string) string {
	// ドットを含む場合やパッケージ名がある場合は、アンダースコアに置換
	sanitized := strings.ReplaceAll(string(id), ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")

	// 特殊文字が含まれている場合は先に処理
	if matched, _ := regexp.MatchString(`[^a-zA-Z0-9_]`, sanitized); matched {
		// 英数字とアンダースコア以外を除去
		reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
		sanitized = reg.ReplaceAllString(sanitized, "_")
	}

	// 数字で始まる場合はプレフィックスを追加
	if matched, _ := regexp.MatchString(`^[0-9]`, sanitized); matched {
		return fmt.Sprintf("node_%s", sanitized)
	}

	// 予約語チェック（大文字小文字を区別しない）
	// パッケージ名を含む場合は最後の部分（型名）のみをチェック
	parts := strings.Split(sanitized, "_")
	lastPart := parts[len(parts)-1]
	if mermaidReservedWords[strings.ToLower(lastPart)] {
		return fmt.Sprintf("node_%s", sanitized)
	}

	// 全体が予約語の場合もチェック
	if mermaidReservedWords[strings.ToLower(sanitized)] {
		return fmt.Sprintf("node_%s", sanitized)
	}

	return sanitized
}

// escapeNodeLabel はノードラベルをmermaidで安全に表示できるように変換する
func escapeNodeLabel(label string) string {
	// HTMLエスケープ
	label = strings.ReplaceAll(label, "&", "&amp;")
	label = strings.ReplaceAll(label, "<", "&lt;")
	label = strings.ReplaceAll(label, ">", "&gt;")
	label = strings.ReplaceAll(label, "\"", "&quot;")
	label = strings.ReplaceAll(label, "'", "&#39;")

	return label
}

// nodeWithStability は不安定度情報を含むノード
type nodeWithStability struct {
	ID          graph.NodeID
	Name        string
	Kind        graph.NodeKind
	Package     string
	Instability float64
	SafeID      string
}

func GenerateMermaid(g *graph.DependencyGraph, stability *graph.StabilityResult) string {

	// パッケージごとにノードをグループ化
	packageNodes := make(map[string][]nodeWithStability)

	for id, n := range g.Nodes {
		inst := 0.0
		if s, ok := stability.NodeStabilities[id]; ok {
			inst = s.Instability
		}

		node := nodeWithStability{
			ID:          id,
			Name:        n.Name,
			Kind:        n.Kind,
			Package:     n.Package,
			Instability: inst,
			SafeID:      sanitizeNodeID(string(id)),
		}

		packageNodes[n.Package] = append(packageNodes[n.Package], node)
	}

	// 各パッケージ内でノードを不安定度降順でソート
	for pkg := range packageNodes {
		sort.Slice(packageNodes[pkg], func(i, j int) bool {
			return packageNodes[pkg][i].Instability > packageNodes[pkg][j].Instability
		})
	}

	out := "graph TD\n"

	// パッケージ名をソートして一貫した順序で出力
	var packages []string
	for pkg := range packageNodes {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	// ノードIDのマッピングを作成
	idMapping := make(map[graph.NodeID]string)

	// パッケージごとにサブグラフを作成
	for _, pkg := range packages {
		nodes := packageNodes[pkg]
		if len(nodes) == 0 {
			continue
		}

		// パッケージノードかどうかをチェック
		isPackageNode := len(nodes) == 1 && nodes[0].Kind == graph.NodePackage

		if !isPackageNode {
			// 通常のパッケージ（構造体、関数、インターフェースを含む）
			safePkgName := sanitizeNodeID(pkg)
			out += fmt.Sprintf("    subgraph %s[\"%s\"]\n", safePkgName, escapeNodeLabel(pkg))

			for _, n := range nodes {
				idMapping[n.ID] = n.SafeID
				escapedName := escapeNodeLabel(n.Name)
				nodeShape := getNodeShape(n.Kind)
				out += fmt.Sprintf("        %s%s\n", n.SafeID, nodeShape(escapedName, n.Instability))
			}

			out += "    end\n"
		} else {
			// パッケージノード（パッケージ間依存関係用）
			n := nodes[0]
			idMapping[n.ID] = n.SafeID
			escapedName := escapeNodeLabel(n.Name)
			nodeShape := getNodeShape(n.Kind)
			out += fmt.Sprintf("    %s%s\n", n.SafeID, nodeShape(escapedName, n.Instability))
		}
	}

	// エッジ定義
	for from, tos := range g.Edges {
		safeFromID := idMapping[from]
		if safeFromID == "" {
			safeFromID = sanitizeNodeID(string(from))
		}

		for to := range tos {
			safeToID := idMapping[to]
			if safeToID == "" {
				safeToID = sanitizeNodeID(string(to))
			}
			out += fmt.Sprintf("    %s --> %s\n", safeFromID, safeToID)
		}
	}

	// スタイル定義を追加
	out += generateStyles()

	// ノードにスタイルクラスを適用
	out += applyNodeStyles(packageNodes)

	return out
}

// getNodeShape はノードの種類に応じた形状を返す関数を返す
func getNodeShape(kind graph.NodeKind) func(string, float64) string {
	switch kind {
	case graph.NodeStruct:
		// 構造体: 長方形
		return func(name string, instability float64) string {
			return fmt.Sprintf("[%s<br>不安定度:%.2f]", name, instability)
		}
	case graph.NodeInterface:
		// インターフェース: 菱形
		return func(name string, instability float64) string {
			return fmt.Sprintf("{%s<br>不安定度:%.2f}", name, instability)
		}
	case graph.NodeFunc:
		// 関数: 角丸長方形
		return func(name string, instability float64) string {
			return fmt.Sprintf("(%s<br>不安定度:%.2f)", name, instability)
		}
	case graph.NodePackage:
		// パッケージ: 六角形
		return func(name string, instability float64) string {
			return fmt.Sprintf("{{%s<br>不安定度:%.2f}}", name, instability)
		}
	default:
		// デフォルト: 長方形
		return func(name string, instability float64) string {
			return fmt.Sprintf("[%s<br>不安定度:%.2f]", name, instability)
		}
	}
}

// generateStyles はMermaidのスタイル定義を生成
func generateStyles() string {
	return `
    %% スタイル定義
    classDef structStyle fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef interfaceStyle fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef funcStyle fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef packageStyle fill:#fff3e0,stroke:#e65100,stroke-width:3px
`
}

// applyNodeStyles はノードにスタイルクラスを適用
func applyNodeStyles(packageNodes map[string][]nodeWithStability) string {
	var out string

	for _, nodes := range packageNodes {
		for _, node := range nodes {
			var styleClass string
			switch node.Kind {
			case graph.NodeStruct:
				styleClass = "structStyle"
			case graph.NodeInterface:
				styleClass = "interfaceStyle"
			case graph.NodeFunc:
				styleClass = "funcStyle"
			case graph.NodePackage:
				styleClass = "packageStyle"
			default:
				styleClass = "structStyle" // デフォルト
			}
			out += fmt.Sprintf("    class %s %s\n", node.SafeID, styleClass)
		}
	}

	return out
}
