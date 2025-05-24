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

func GenerateMermaid(g *graph.DependencyGraph, stability *graph.StabilityResult) string {
	type nodeWithStability struct {
		ID          graph.NodeID
		Name        string
		Instability float64
		SafeID      string
	}
	var nodes []nodeWithStability
	for id, n := range g.Nodes {
		inst := 0.0
		if s, ok := stability.NodeStabilities[id]; ok {
			inst = s.Instability
		}
		nodes = append(nodes, nodeWithStability{
			ID:          id,
			Name:        n.Name,
			Instability: inst,
			SafeID:      sanitizeNodeID(string(id)),
		})
	}
	// 不安定度降順でソート
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Instability > nodes[j].Instability
	})

	out := "graph TD\n"

	// ノード定義
	for _, n := range nodes {
		escapedName := escapeNodeLabel(n.Name)
		out += fmt.Sprintf("    %s[\"%s<br>不安定度:%.2f\"]\n", n.SafeID, escapedName, n.Instability)
	}

	// エッジ定義
	// ノードIDのマッピングを作成
	idMapping := make(map[graph.NodeID]string)
	for _, n := range nodes {
		idMapping[n.ID] = n.SafeID
	}

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
	return out
}
