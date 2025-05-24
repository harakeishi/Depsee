package analyzer

// Analyzer はコード解析を行うインターフェース
type Analyzer interface {
	AnalyzeDir(dir string) (*AnalysisResult, error)
}
