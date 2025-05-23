# 全体構成の設計

## ディレクトリ構成案

```
depsee/
├── cmd/                # CLIエントリポイント
│   └── depsee/         # depseeコマンド本体
│       └── main.go
├── internal/           # 内部ロジック（パッケージ分割）
│   ├── analyzer/       # 解析ロジック（AST解析・依存抽出）
│   │   ├── analyzer.go
│   │   └── types.go
│   ├── graph/          # 依存グラフ構築・安定度計算
│   │   ├── graph.go
│   │   └── stability.go
│   └── output/         # Mermaid記法など出力系
│       └── mermaid.go
├── testdata/           # サンプルGoコード（テスト用）
├── go.mod
├── go.sum
└── README.md
```

---

## 主要ファイル・役割

- `cmd/depsee/main.go`  
  CLIエントリポイント。引数パース、サブコマンド分岐、各機能呼び出し。

- `internal/analyzer/analyzer.go`  
  ディレクトリ走査、GoファイルのAST解析、構造体・関数・インターフェース抽出。

- `internal/analyzer/types.go`  
  解析対象（構造体・関数・インターフェース）のデータ構造定義。

- `internal/graph/graph.go`  
  依存グラフの内部表現、依存関係の追加・取得。

- `internal/graph/stability.go`  
  依存数・非依存数・安定度の計算ロジック。

- `internal/output/mermaid.go`  
  依存グラフをMermaid記法に変換し出力。

- `testdata/`  
  テスト・検証用のサンプルGoコード。

---

## 次のステップ

1. go.mod初期化・プロジェクト雛形作成
2. `cmd/depsee/main.go` のCLIエントリポイント実装
3. `internal/analyzer/`のAST解析・抽出ロジック実装

---

今後も各ステップで決まった内容はmdで記録します。 