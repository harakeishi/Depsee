# 全体構成の設計

## ディレクトリ構成案

```
depsee/
├── cmd/                # CLIエントリポイント
│   ├── analyze.go      # analyzeコマンド
│   ├── root.go         # ルートコマンド
│   ├── version.go      # versionコマンド
│   └── main_test.go    # テスト
├── internal/           # 内部ロジック（パッケージ分割）
│   ├── analyzer/       # 解析ロジック（AST解析・依存抽出）
│   │   ├── analyzer.go
│   │   ├── analyzer_test.go
│   │   ├── interfaces.go
│   │   ├── types.go
│   │   └── type_resolver.go
│   ├── errors/         # エラーハンドリング
│   │   └── errors.go
│   ├── graph/          # 依存グラフ構築・安定度計算
│   │   ├── builder.go
│   │   ├── extractor.go
│   │   ├── graph.go
│   │   ├── graph_test.go
│   │   ├── stability.go
│   │   └── stability_test.go
│   ├── logger/         # ログ出力
│   │   └── logger.go
│   ├── output/         # Mermaid記法など出力系
│   │   ├── generator.go
│   │   └── mermaid.go
│   └── utils/          # ユーティリティ関数
├── pkg/depsee/         # パブリックAPI
│   ├── depsee.go       # メインAPI
│   └── depsee_test.go  # テスト
├── testdata/           # サンプルGoコード（テスト用）
│   └── sample/
├── docs/               # 設計ドキュメント
├── main.go             # エントリポイント
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE
└── README.md
```

---

## 主要ファイル・役割

- `main.go`  
  CLIエントリポイント。cmdパッケージのExecute()を呼び出し。

- `cmd/root.go`  
  ルートコマンドの定義。グローバルフラグとログ設定の初期化。

- `cmd/analyze.go`  
  analyzeサブコマンドの実装。引数パース、設定構築、解析実行。

- `cmd/version.go`  
  versionサブコマンドの実装。バージョン情報の表示。

- `pkg/depsee/depsee.go`  
  メインAPIの実装。設定に基づく解析の実行と結果表示。

- `internal/analyzer/analyzer.go`  
  ディレクトリ走査、GoファイルのAST解析、構造体・関数・インターフェース抽出。

- `internal/analyzer/interfaces.go`  
  解析機能のインターフェース定義。

- `internal/analyzer/types.go`  
  解析対象（構造体・関数・インターフェース）のデータ構造定義。

- `internal/analyzer/type_resolver.go`  
  型解決とパッケージ情報の抽出ロジック。

- `internal/errors/errors.go`  
  カスタムエラー型とエラー収集機能。

- `internal/graph/graph.go`  
  依存グラフの内部表現、依存関係の追加・取得。

- `internal/graph/builder.go`  
  依存グラフ構築のメインロジック。

- `internal/graph/extractor.go`  
  依存関係抽出の戦略パターン実装。

- `internal/graph/stability.go`  
  依存数・非依存数・安定度の計算ロジック。

- `internal/logger/logger.go`  
  構造化ログ出力機能。

- `internal/output/generator.go`  
  出力生成のインターフェースと実装。

- `internal/output/mermaid.go`  
  依存グラフをMermaid記法に変換し出力。

- `testdata/sample/`  
  テスト・検証用のサンプルGoコード。

---

## 関連ドキュメント

- [要件定義](requirements.md) - プロジェクト全体の要件と仕様
- [静的解析設計](design_static_analysis.md) - Goコードの解析機能
- [依存関係解析設計](design_dependency_analysis.md) - 依存関係抽出機能
- [不安定度解析設計](design_stability_analysis.md) - 不安定度計算機能
- [Mermaid出力設計](design_mermaid_output.md) - 相関図生成機能
- [CLI機能設計](design_cli.md) - コマンドラインインターフェース
- [ログ機能設計](design_logging.md) - ログ出力機能
- [エラーハンドリング設計](design_error_handling.md) - エラー処理機能

---

## 実装状況

✅ **完了済み**
1. go.mod初期化・プロジェクト雛形作成
2. `cmd/depsee/main.go` のCLIエントリポイント実装
3. `internal/analyzer/`のAST解析・抽出ロジック実装
4. `internal/graph/`の依存グラフ構築・安定度計算実装
5. `internal/output/`のMermaid出力実装
6. `internal/errors/`のエラーハンドリング実装
7. 包括的なテストスイートの実装

---

今後も機能拡張や改善内容はドキュメントで記録します。 
