# コア機能設計：依存関係の抽出

## 目的

- 構造体・インターフェース・関数（メソッド含む）間の依存関係（利用・参照）を静的解析により抽出し、依存グラフを構築する。

---

## 要件

- 構造体・インターフェース・関数（メソッド）は静的解析で抽出済みの情報を利用
- 依存関係の種類（例）：
    - フィールド型として他の構造体/インターフェースを参照
    - メソッドや関数の引数・戻り値型として参照
    - メソッド・関数内で他の型や関数を利用
    - インターフェースの実装関係
- 依存元・依存先の情報をグラフ構造で保持
- 依存関係の重複は排除
- 依存関係の抽出はパッケージ内・パッケージ横断の両方に対応

---

## 入力

- 静的解析で得られた構造体・インターフェース・関数（メソッド含む）の情報

---

## 出力

- 依存関係グラフ（ノード：構造体/インターフェース/関数、エッジ：依存関係）

---

## 主な処理フロー

1. **フィールド依存の抽出**
    - 構造体の各フィールド型を調査し、他の構造体・インターフェース型への依存を抽出

2. **メソッド・関数シグネチャ依存の抽出**
    - 引数・戻り値の型を調査し、他の型への依存を抽出

3. **メソッド・関数本体依存の抽出**
    - 関数・メソッド内で利用されている型・関数・メソッドをASTから抽出

4. **インターフェース実装依存の抽出**
    - 構造体がインターフェースを実装している場合、その依存を抽出

5. **依存グラフへの格納**
    - 依存元・依存先のノードをエッジで接続
    - 重複エッジは排除
    - 戦略パターンを使用して依存関係抽出ロジックを分離

---

## 想定データ構造（例：graph.go）

```go
type NodeKind int

const (
    NodeStruct NodeKind = iota
    NodeInterface
    NodeFunc
)

type NodeID string // 例: "package.StructName" など

type Node struct {
    ID      NodeID
    Kind    NodeKind
    Name    string
    Package string
}

type Edge struct {
    From NodeID
    To   NodeID
    // 依存の種類（フィールド/シグネチャ/本体/実装）も必要に応じて
}

type DependencyGraph struct {
    Nodes map[NodeID]*Node
    Edges map[NodeID]map[NodeID]struct{} // From→Toの多重辺排除
}

// 依存関係抽出の戦略パターン
type DependencyExtractor interface {
    Extract(result *analyzer.AnalysisResult, graph *DependencyGraph)
}

// フィールド依存抽出
type FieldDependencyExtractor struct {
    typeResolver *analyzer.TypeResolver
}

// シグネチャ依存抽出
type SignatureDependencyExtractor struct{}

// 本体呼び出し依存抽出
type BodyCallDependencyExtractor struct{}
```

---

## 拡張ポイント

- 依存関係の種類（フィールド依存、シグネチャ依存、実装依存など）をエッジ属性として保持
- パッケージ間依存の可視化
- ジェネリクスや埋め込み型への対応 
