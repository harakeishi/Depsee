# depsee

Goコードの構造体・関数・インターフェースの依存関係を可視化し、不安定度（変更容易度）をMermaid記法で出力するCLIツール

## 特徴

- 🔍 **静的解析**: Goコードを解析して構造体・関数・インターフェースを抽出
- 📊 **依存関係可視化**: 要素間の依存関係をグラフ構造で表現
- 📦 **パッケージ間依存関係**: 同リポジトリ内のパッケージ間依存関係を解析（オプション）
- 🎯 **パッケージフィルタリング**: 指定されたパッケージのみを解析対象とする機能
- 📈 **不安定度計算**: SOLID原則に基づく不安定度指標の算出
- 🎨 **Mermaid出力**: 相関図をMermaid記法で生成
- 🛠️ **高品質設計**: SOLIDの原則に準拠した拡張可能なアーキテクチャ

## インストール

```bash
go install github.com/harakeishi/depsee/cmd/depsee@latest
```

## 使用方法

### 基本的な使用例

```bash
# プロジェクトの解析
depsee analyze ./path/to/your/project

# パッケージ間依存関係を含む解析
depsee --include-package-deps analyze ./path/to/your/project

# 特定のパッケージのみを解析
depsee analyze --target-packages main ./path/to/your/project

# 複数のパッケージを解析
depsee analyze --target-packages main,cmd,pkg ./path/to/your/project

# バージョン表示
depsee -version

# デバッグログ付きで実行
depsee -log-level debug analyze ./path/to/project

# JSONログフォーマットで実行
depsee -log-format json analyze ./path/to/project
```

### パッケージフィルタリング

`--target-packages` オプションを使用すると、指定されたパッケージのみを解析対象とできます：

```bash
# mainパッケージのみを解析
depsee analyze --target-packages main ./your-project

# mainとcmdパッケージのみを解析
depsee analyze --target-packages main,cmd ./your-project

# 複数パッケージの解析（スペースを含む場合はクォートで囲む）
depsee analyze --target-packages "main, cmd, internal/service" ./your-project
```

この機能により、以下のメリットがあります：
- **大規模プロジェクトでの効率的な解析**: 関心のあるパッケージのみに焦点を当てることができます
- **段階的な解析**: パッケージごとに依存関係を段階的に確認できます
- **パフォーマンス向上**: 解析対象を絞ることで処理時間を短縮できます

### パッケージ間依存関係解析

`--include-package-deps` オプションを使用すると、同リポジトリ内のパッケージ間の依存関係も解析できます：

```bash
# パッケージ間依存関係を含む解析
depsee --include-package-deps analyze ./multi-package-project

# パッケージフィルタリングと組み合わせて使用
depsee analyze --target-packages main,cmd --include-package-deps ./multi-package-project
```

この機能により、以下が追加で解析されます：
- パッケージノード（`package:パッケージ名`）
- パッケージ間の依存関係（import文に基づく）
- 標準ライブラリは除外され、同リポジトリ内のパッケージのみが対象

### 出力例

```
[info] 構造体一覧:
  - User (package: sample, file: testdata/sample/user.go)
      * メソッド: UpdateProfile
      * メソッド: AddPost
  - Profile (package: sample, file: testdata/sample/user.go)
  - Post (package: sample, file: testdata/sample/user.go)
  - UserSettings (package: sample, file: testdata/sample/user.go)

[info] インターフェース一覧:
  - UserService (package: sample, file: testdata/sample/user.go)

[info] 関数一覧:
  - CreateUser (package: sample, file: testdata/sample/user.go)
  - GetUserPosts (package: sample, file: testdata/sample/user.go)

[info] 依存グラフ ノード:
  - sample.User (User)
  - sample.Profile (Profile)
  - sample.Post (Post)
  - sample.UserSettings (UserSettings)
  - sample.UserService (UserService)
  - sample.CreateUser (CreateUser)
  - sample.GetUserPosts (GetUserPosts)

[info] ノード不安定度:
  sample.User: 依存数=3, 非依存数=3, 不安定度=0.50
  sample.Post: 依存数=1, 非依存数=2, 不安定度=0.33
  sample.UserService: 依存数=0, 非依存数=0, 不安定度=1.00
  sample.CreateUser: 依存数=1, 非依存数=0, 不安定度=1.00

[info] Mermaid相関図:
graph TD
    sample.UserService["UserService<br>不安定度:1.00"]
    sample.CreateUser["CreateUser<br>不安定度:1.00"]
    sample.GetUserPosts["GetUserPosts<br>不安定度:1.00"]
    sample.User["User<br>不安定度:0.50"]
    sample.Post["Post<br>不安定度:0.33"]
    sample.Profile["Profile<br>不安定度:0.00"]
    sample.UserSettings["UserSettings<br>不安定度:0.00"]
    sample.User --> sample.Profile
    sample.User --> sample.Post
    sample.User --> sample.UserSettings
    sample.Post --> sample.User
    sample.CreateUser --> sample.User
    sample.GetUserPosts --> sample.User
    sample.GetUserPosts --> sample.Post
```

## ディレクトリ構成

```
depsee/
├── cmd/depsee/           # CLIエントリポイント
├── internal/
│   ├── analyzer/         # 静的解析ロジック
│   ├── cli/              # CLIロジック
│   ├── errors/           # エラーハンドリング
│   ├── graph/            # 依存グラフ・安定度算出
│   ├── logger/           # ログ機能
│   └── output/           # Mermaid出力
├── testdata/sample/      # サンプルGoコード・テスト用
└── docs/                 # 設計ドキュメント
```

## 開発

### ビルド

```bash
go build -o depsee cmd/depsee/main.go
```

### テスト

```bash
go test ./...
```

### 開発用実行

```bash
go run cmd/depsee/main.go analyze ./testdata/sample
```

## アーキテクチャ

このプロジェクトはSOLIDの原則に基づいて設計されており、以下の特徴があります：

- **単一責任の原則**: 各パッケージが明確な責任を持つ
- **依存関係逆転の原則**: インターフェースを通じた疎結合
- **戦略パターン**: 依存関係抽出ロジックの柔軟な拡張
- **依存性注入**: 高いテスタビリティ

## ドキュメント

詳細な設計・仕様については `docs/` ディレクトリの設計ドキュメントを参照してください：

- [全体設計](docs/design.md)
- [要件定義](docs/requirements.md)
- [静的解析設計](docs/design_static_analysis.md)
- [依存関係解析設計](docs/design_dependency_analysis.md)
- [不安定度解析設計](docs/design_stability_analysis.md)
- [Mermaid出力設計](docs/design_mermaid_output.md)
- [CLI機能設計](docs/design_cli.md)
- [ログ機能設計](docs/design_logging.md)
- [エラーハンドリング設計](docs/design_error_handling.md)

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルを参照してください。
