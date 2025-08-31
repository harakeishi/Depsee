package analyzer

// Analyzer はコード解析を行うインターフェース
type Analyzer interface {
	SetFilters(filters Filters)        // 解析フィルタを設定する
	ListTartgetFiles(dir string) error // 解析対象のGoファイルをリストアップする
	Analyze() error                    // 解析を行う
	ExportResult() *Result             // 解析結果をエクスポートする
}

/*
- Analyzerの責務
	- 指定ディレクトリ配下のGoファイルを再帰的に探索し、構造体・インターフェース・関数を抽出する
	- パッケージが指定されている場合、指定されたパッケージのみを解析対象とする
	- 構造体、インターフェース、関数の意依存関係を抽出する
*/
