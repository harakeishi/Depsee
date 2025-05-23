# コア機能設計：Goコードの静的解析

## 目的

- 指定ディレクトリ配下のGoファイルを再帰的に探索し、  
  各ファイルから構造体・関数・インターフェースを抽出する。

---

## 要件

- ディレクトリを再帰的に探索し、`.go`ファイルのみを対象とする
- テストファイル（`*_test.go`）は除外
- Go標準パッケージ（`go/parser`, `go/ast`）を利用
- 構造体（type ... struct）、関数（func ...）、インターフェース（type ... interface）の一覧を取得
- **メソッドは構造体（レシーバ型）に内包し、構造体単位で安定度を算出・可視化する**
- 取得した情報は後続の依存関係解析・グラフ生成で利用できるようにデータ構造化
- **同一リポジトリ内のimport先パッケージも再帰的に解析対象とする（オプションで有効化）**

---

## 入力

- 解析対象ディレクトリのパス
- （オプション）`--with-local-imports` フラグ

---

## 出力

- 構造体、関数、インターフェースの一覧（データ構造として保持）

---

## 主な処理フロー

1. **ディレクトリ再帰探索**
    - 指定ディレクトリ配下の全`.go`ファイルをリストアップ
    - `filepath.WalkDir`等を利用
    - **import文を抽出し、同一リポジトリ内のパッケージも再帰的に解析（オプション）**
    - 既に解析済みのパッケージは再解析しない

2. **Goファイルのパース**
    - `go/parser`でASTを生成
    - `go/ast`でノードを走査

3. **構造体・関数・インターフェースの抽出**
    - `ast.GenDecl`から`type`宣言を抽出し、`struct`/`interface`を判定
    - `ast.FuncDecl`から関数・メソッドを抽出
    - **メソッドはレシーバ型（構造体）に紐づけて格納**

4. **データ構造への格納**
    - 構造体、関数、インターフェースごとに情報を格納
    - 名前、パッケージ名、ファイルパス、位置情報などを保持
    - **構造体（StructInfo）は自身に属するメソッド（FuncInfo）のリストを持つ**

---

## 想定データ構造（例：types.go）

```go
type StructInfo struct {
    Name        string
    Package     string
    File        string
    Position    token.Position
    Methods     []FuncInfo // この構造体に属するメソッド
    Fields      []FieldInfo // フィールド情報
}

type InterfaceInfo struct {
    Name        string
    Package     string
    File        string
    Position    token.Position
    Methods     []FuncInfo // インターフェースのメソッド
}

type FuncInfo struct {
    Name        string
    Receiver    string // メソッドの場合、レシーバ型名
    Package     string
    File        string
    Position    token.Position
}
```

---

## 拡張ポイント

- フィールド・メソッドの詳細情報も必要に応じて格納
- パッケージ単位での集約
- ジェネリクスや埋め込み型への対応
- **ローカルimport再帰解析の効率化・循環参照対策** 