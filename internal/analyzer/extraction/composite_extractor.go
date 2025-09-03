package extraction

import (
	"go/ast"
	"go/token"
)

// CompositeExtractor combines multiple extraction strategies
type CompositeExtractor struct {
	strategies []ExtractionStrategy
}

// NewCompositeExtractor creates a new composite extractor
func NewCompositeExtractor(strategies ...ExtractionStrategy) *CompositeExtractor {
	return &CompositeExtractor{
		strategies: strategies,
	}
}

// ExtractDependencies runs all strategies and combines results
func (e *CompositeExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var allDependencies []DependencyInfo
	
	for _, strategy := range e.strategies {
		deps, err := strategy.ExtractDependencies(file, fset, packageName)
		if err != nil {
			return nil, err
		}
		allDependencies = append(allDependencies, deps...)
	}
	
	return allDependencies, nil
}

// Name returns the strategy name
func (e *CompositeExtractor) Name() string {
	return "CompositeExtractor"
}

// AddStrategy adds a new extraction strategy
func (e *CompositeExtractor) AddStrategy(strategy ExtractionStrategy) {
	e.strategies = append(e.strategies, strategy)
}

// RemoveStrategy removes a strategy by name
func (e *CompositeExtractor) RemoveStrategy(name string) {
	var newStrategies []ExtractionStrategy
	for _, s := range e.strategies {
		if s.Name() != name {
			newStrategies = append(newStrategies, s)
		}
	}
	e.strategies = newStrategies
}