package analyzer

// Analyzer は静的解析を行うサービス
type Analyzer struct {
	typeResolver *TypeResolver
}

// New は新しいAnalyzerインスタンスを作成
func New() *Analyzer {
	return &Analyzer{
		typeResolver: NewTypeResolver(),
	}
}

// AnalyzeDir は指定ディレクトリ配下のGoファイルを解析
func (a *Analyzer) AnalyzeDir(dir string) (*AnalysisResult, error) {
	return AnalyzeDir(dir)
}
