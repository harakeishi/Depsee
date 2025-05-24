# ログ機能設計

## 目的

- 解析処理の進行状況とエラー情報を構造化ログで記録
- デバッグ時の詳細情報提供
- 本番環境での適切なログレベル制御

---

## 要件

- 構造化ログ（JSON/テキスト形式）の対応
- ログレベルによる出力制御（debug, info, warn, error）
- 設定可能な出力先（標準エラー出力等）
- パフォーマンスを考慮した効率的なログ出力

---

## ログレベル

### Debug
- 詳細なデバッグ情報
- ファイル発見、AST解析の詳細など
- 開発・デバッグ時のみ使用

### Info
- 一般的な処理状況
- 解析開始・完了、ファイル数など
- 通常運用時の標準レベル

### Warn
- 警告レベルの問題
- パースエラー、部分的な失敗など
- 処理は継続するが注意が必要

### Error
- エラーレベルの問題
- 解析失敗、重大なエラーなど
- 処理の中断を伴う問題

---

## ログフォーマット

### テキスト形式
```
2024/01/01 12:00:00 INFO 解析開始 target_dir=/path/to/project
2024/01/01 12:00:01 DEBUG Goファイル発見 file=/path/to/file.go
```

### JSON形式
```json
{"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"解析開始","target_dir":"/path/to/project"}
{"time":"2024-01-01T12:00:01Z","level":"DEBUG","msg":"Goファイル発見","file":"/path/to/file.go"}
```

---

## 実装データ構造

```go
type LogLevel string

const (
    LevelDebug LogLevel = "debug"
    LevelInfo  LogLevel = "info"
    LevelWarn  LogLevel = "warn"
    LevelError LogLevel = "error"
)

type Config struct {
    Level  LogLevel
    Format string // "text" or "json"
    Output io.Writer
}

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

---

## 使用例

```go
// ログ設定の初期化
logger.Init(logger.Config{
    Level:  logger.LevelInfo,
    Format: "text",
    Output: os.Stderr,
})

// ログ出力
logger.Info("解析開始", "target_dir", "/path/to/project")
logger.Debug("Goファイル発見", "file", "/path/to/file.go")
logger.Warn("パースエラー", "file", "/path/to/file.go", "error", err)
```

---

## 拡張ポイント

- ログローテーション機能
- 複数出力先への同時出力
- メトリクス情報の記録
- ログフィルタリング機能 
