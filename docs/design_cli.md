# CLI機能設計

## 目的

- ユーザーフレンドリーなコマンドラインインターフェースの提供
- 依存性注入による高いテスタビリティの実現
- 設定可能なオプションによる柔軟な実行制御

---

## 要件

- サブコマンド形式のインターフェース（`analyze`コマンド）
- バージョン表示機能
- ログレベル・フォーマットの設定
- エラーハンドリングとユーザーフレンドリーなメッセージ
- 依存性注入によるテスト容易性

---

## コマンド仕様

### 基本構文
```
depsee [options] <command> [arguments]
```

### オプション
- `-version`: バージョン情報を表示
- `-log-level <level>`: ログレベル指定 (debug, info, warn, error)
- `-log-format <format>`: ログフォーマット指定 (text, json)

### サブコマンド
- `analyze <target_dir>`: 指定ディレクトリの解析を実行

---

## 実行フロー

1. **引数解析**
   - コマンドラインオプションの解析
   - サブコマンドと引数の検証

2. **設定初期化**
   - ログ設定の初期化
   - 依存関係の注入

3. **コマンド実行**
   - 解析処理の実行
   - 結果の表示

4. **結果出力**
   - 構造体・インターフェース・関数の一覧表示
   - 依存グラフの表示
   - 不安定度情報の表示
   - Mermaid相関図の出力

---

## 実装データ構造

```go
type Config struct {
    ShowVersion bool
    LogLevel    string
    LogFormat   string
    TargetDir   string
}

type CLI struct {
    analyzer  analyzer.Analyzer
    grapher   graph.GraphBuilder
    outputter output.OutputGenerator
    logger    logger.Logger
}
```

---

## 依存性注入

### デフォルト構成
```go
func NewCLI() *CLI {
    return NewCLIWithDependencies(
        analyzer.New(),
        graph.NewBuilder(),
        output.NewGenerator(),
        logger.NewLogger(defaultConfig),
    )
}
```

### テスト用構成
```go
func NewCLIWithDependencies(
    analyzer analyzer.Analyzer,
    grapher graph.GraphBuilder,
    outputter output.OutputGenerator,
    logger logger.Logger,
) *CLI
```

---

## 出力形式

### 解析結果表示
```
[info] 構造体一覧:
  - User (package: main, file: /path/to/user.go)
      * メソッド: GetName
      * メソッド: SetEmail

[info] インターフェース一覧:
  - UserService (package: main, file: /path/to/service.go)

[info] 関数一覧:
  - CreateUser (package: main, file: /path/to/user.go)
```

### 依存グラフ表示
```
[info] 依存グラフ ノード:
  - main.User (User)
  - main.UserService (UserService)

[info] 依存グラフ エッジ:
  main.CreateUser --> main.User
```

### 不安定度表示
```
[info] ノード不安定度:
  main.User: 依存数=0, 非依存数=2, 不安定度=0.00
  main.CreateUser: 依存数=1, 非依存数=0, 不安定度=1.00
```

---

## エラーハンドリング

### 引数エラー
- 不正なコマンド形式
- 存在しないディレクトリ指定
- 無効なオプション値

### 実行エラー
- 解析処理の失敗
- ファイルアクセスエラー
- パースエラー

---

## 拡張ポイント

- 設定ファイルからの設定読み込み
- 出力先の指定（ファイル出力）
- 除外パターンの指定
- 並列処理オプション
- プログレスバー表示 
