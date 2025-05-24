# depsee

[![CI](https://github.com/harakeishi/Depsee/actions/workflows/ci.yml/badge.svg)](https://github.com/harakeishi/Depsee/actions/workflows/ci.yml)
[![Auto Release](https://github.com/harakeishi/Depsee/actions/workflows/auto-release.yml/badge.svg)](https://github.com/harakeishi/Depsee/actions/workflows/auto-release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/harakeishi/depsee)](https://goreportcard.com/report/github.com/harakeishi/depsee)
[![codecov](https://codecov.io/gh/harakeishi/Depsee/branch/main/graph/badge.svg)](https://codecov.io/gh/harakeishi/Depsee)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/harakeishi/Depsee)](https://github.com/harakeishi/Depsee/releases/latest)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io%2Fharakeishi%2Fdepsee-blue)](https://github.com/harakeishi/Depsee/pkgs/container/depsee)

Goコードの構造体・関数・インターフェースの依存関係を可視化し、不安定度（変更容易度）をMermaid記法で出力するCLIツール

## 特徴

- 🔍 **静的解析**: Goコードを解析して構造体・関数・インターフェースを抽出
- 📊 **依存関係可視化**: 要素間の依存関係をグラフ構造で表現
- 📦 **パッケージ間依存関係**: 同リポジトリ内のパッケージ間依存関係を解析（オプション）
- 🎯 **パッケージフィルタリング**: 指定されたパッケージのみを解析対象とする機能
- 🚫 **除外機能**: 指定されたパッケージやディレクトリを解析対象から除外する機能
- 📈 **不安定度計算**: SOLID原則に基づく不安定度指標の算出
- 🎨 **Mermaid出力**: 相関図をMermaid記法で生成
- 🛠️ **高品質設計**: SOLIDの原則に準拠した拡張可能なアーキテクチャ

## インストール

### Go (推奨)

```bash
go install github.com/harakeishi/depsee@latest
```

### Homebrew

```bash
# タップを追加
brew tap harakeishi/tap

# インストール
brew install depsee
```

### Docker

```bash
# 最新版を使用
docker pull ghcr.io/harakeishi/depsee:latest

# 特定バージョンを使用
docker pull ghcr.io/harakeishi/depsee:v1.0.0
```

### バイナリダウンロード

[GitHub Releases](https://github.com/harakeishi/Depsee/releases/latest)から、お使いのプラットフォーム向けのバイナリをダウンロードできます：

- **Linux**: `depsee_Linux_x86_64.tar.gz`
- **macOS (Intel)**: `depsee_Darwin_x86_64.tar.gz`
- **macOS (Apple Silicon)**: `depsee_Darwin_arm64.tar.gz`
- **Windows**: `depsee_Windows_x86_64.zip`

### ソースからビルド

```bash
git clone https://github.com/harakeishi/Depsee.git
cd Depsee
go build -o depsee .
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

# 特定のパッケージを除外
depsee analyze --exclude-packages test,mock ./path/to/your/project

# 特定のディレクトリを除外
depsee analyze --exclude-dirs testdata,vendor ./path/to/your/project

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

### 除外機能

`--exclude-packages` と `--exclude-dirs` オプションを使用すると、指定されたパッケージやディレクトリを解析対象から除外できます：

```bash
# testパッケージを除外
depsee analyze --exclude-packages test ./your-project

# 複数のパッケージを除外
depsee analyze --exclude-packages test,mock,vendor ./your-project

# testdataディレクトリを除外
depsee analyze --exclude-dirs testdata ./your-project

# 複数のディレクトリを除外
depsee analyze --exclude-dirs testdata,vendor,third_party ./your-project

# パッケージとディレクトリの両方を除外
depsee analyze --exclude-packages test --exclude-dirs vendor ./your-project

# フィルタリングと除外を組み合わせ
depsee analyze --target-packages main,cmd --exclude-packages test --exclude-dirs testdata ./your-project
```

この機能により、以下のメリットがあります：
- **不要なコードの除外**: テストコードやベンダーコードなど、解析に不要な部分を除外できます
- **効率的な解析**: 除外により処理時間を短縮し、関心のある部分に集中できます
- **柔軟な設定**: パッケージレベルとディレクトリレベルの両方で除外設定が可能です
- **組み合わせ可能**: target-packagesと組み合わせて、より細かい制御が可能です

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

## CI/CD & デプロイメント

このプロジェクトは完全自動化されたCI/CDパイプラインを採用しており、コードの品質管理から自動リリースまでを自動化しています。

### 🔄 継続的インテグレーション (CI)

**トリガー**: `main`、`develop`ブランチへのプッシュ、および`main`ブランチへのプルリクエスト

**実行内容**:
- **マルチバージョンテスト**: Go 1.21、1.22、1.23での動作確認
- **静的解析**: golangci-lintによるコード品質チェック
- **テスト実行**: レースコンディション検出付きテスト
- **カバレッジ測定**: Codecovへの自動アップロード
- **マルチプラットフォームビルド**: Linux、macOS、Windows向けバイナリ生成

```yaml
# .github/workflows/ci.yml で定義
- Linux (amd64)
- macOS (amd64, arm64)  
- Windows (amd64)
```

### 🚀 自動リリース

**トリガー**: `main`ブランチへのプッシュ（特定のコミットタイプを除く）

**自動バージョニング**:
- **Major版** (`v1.0.0 → v2.0.0`): `feat!:`で始まるコミットまたは`BREAKING CHANGE`を含むコミット
- **Minor版** (`v1.0.0 → v1.1.0`): `feat:`で始まるコミット
- **Patch版** (`v1.0.0 → v1.0.1`): その他のコミット（`fix:`、`refactor:`など）

**除外されるコミット**:
- `chore:`で始まるコミット
- `docs:`で始まるコミット
- `ci:`で始まるコミット
- `[skip release]`または`[skip ci]`を含むコミット

**自動実行される処理**:
1. 最新タグから次のバージョンを自動計算
2. `cmd/root.go`のバージョン変数を自動更新
3. バージョン更新をコミット・プッシュ
4. 新しいタグを作成・プッシュ
5. Go Releaserによるリリース作成

### 📦 リリース成果物

**バイナリ**:
- `depsee_Linux_x86_64.tar.gz`
- `depsee_Darwin_x86_64.tar.gz`
- `depsee_Darwin_arm64.tar.gz`
- `depsee_Windows_x86_64.zip`

**コンテナイメージ**:
- `ghcr.io/harakeishi/depsee:latest`
- `ghcr.io/harakeishi/depsee:v{version}`

**パッケージマネージャー**:
- **Homebrew**: `brew install harakeishi/tap/depsee`
- **Go**: `go install github.com/harakeishi/depsee@latest`

**その他**:
- チェックサムファイル (`checksums.txt`)
- 自動生成されたリリースノート
- CHANGELOGの自動更新

### 🐳 Docker利用

```bash
# 最新版を実行
docker run --rm -v $(pwd):/workspace ghcr.io/harakeishi/depsee:latest analyze /workspace

# 特定バージョンを実行
docker run --rm -v $(pwd):/workspace ghcr.io/harakeishi/depsee:v1.0.0 analyze /workspace
```

### 🔧 手動リリース

自動リリースに加えて、手動でのリリースも可能です：

```bash
# 手動でタグを作成してリリース
git tag v1.2.3
git push origin v1.2.3
```

### 📋 リリース例

**パッチリリース**:
```bash
git commit -m "fix: バグを修正"
git push origin main
# → v1.0.0 → v1.0.1 に自動リリース
```

**マイナーリリース**:
```bash
git commit -m "feat: 新機能を追加"
git push origin main
# → v1.0.0 → v1.1.0 に自動リリース
```

**メジャーリリース**:
```bash
git commit -m "feat!: 破壊的変更を含む新機能"
git push origin main
# → v1.0.0 → v2.0.0 に自動リリース
```

**リリースをスキップ**:
```bash
git commit -m "docs: READMEを更新 [skip release]"
git push origin main
# → リリースされない
```

### 🔍 品質保証

**静的解析設定** (`.golangci.yml`):
- 40以上のlinterを有効化
- プロジェクト固有の設定でコード品質を保証
- テストファイルには緩和されたルールを適用

**テスト戦略**:
- ユニットテスト
- レースコンディション検出
- カバレッジ測定とレポート

**セキュリティ**:
- GitHub Container Registryへの安全な認証
- 最小権限の原則に基づくワークフロー権限設定
- Dockerイメージの脆弱性スキャン

### 📊 監視とメトリクス

- **ビルド状況**: GitHub Actionsのステータスバッジ
- **コードカバレッジ**: Codecovによる可視化
- **リリース履歴**: GitHub Releasesページで確認可能
- **ダウンロード統計**: GitHub Releasesの統計情報

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
