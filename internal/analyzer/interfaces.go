package analyzer

// Analyzer はコード解析を行うインターフェース
type Analyzer interface {
	AnalyzeDir(dir string) (*AnalysisResult, error)
	AnalyzeDirWithPackageFilter(dir string, targetPackages []string) (*AnalysisResult, error)
	AnalyzeDirWithFilters(dir string, targetPackages []string, excludePackages []string, excludeDirs []string) (*AnalysisResult, error)
}
