package errors

import (
	"fmt"
	"strings"
)

// AnalysisError は解析エラーを表す
type AnalysisError struct {
	Path string
	Err  error
}

func (e *AnalysisError) Error() string {
	return fmt.Sprintf("analysis failed for %s: %v", e.Path, e.Err)
}

func (e *AnalysisError) Unwrap() error {
	return e.Err
}

// NewAnalysisError は新しいAnalysisErrorを作成
func NewAnalysisError(path string, err error) *AnalysisError {
	return &AnalysisError{
		Path: path,
		Err:  err,
	}
}

// ParseError はパースエラーを表す
type ParseError struct {
	File string
	Line int
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error in %s at line %d: %v", e.File, e.Line, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// NewParseError は新しいParseErrorを作成
func NewParseError(file string, line int, err error) *ParseError {
	return &ParseError{
		File: file,
		Line: line,
		Err:  err,
	}
}

// ErrorCollector はエラーを収集する
type ErrorCollector struct {
	errors []error
}

// NewErrorCollector は新しいErrorCollectorを作成
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Add はエラーを追加
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// HasErrors はエラーがあるかどうかを返す
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// Errors はエラーのスライスを返す
func (ec *ErrorCollector) Errors() []error {
	return ec.errors
}

// Error はエラーメッセージを返す
func (ec *ErrorCollector) Error() string {
	if len(ec.errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ec.errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("multiple errors occurred:\n%s", strings.Join(messages, "\n"))
}

// ToError はエラーがある場合にErrorCollector自体をerrorとして返す
func (ec *ErrorCollector) ToError() error {
	if ec.HasErrors() {
		return ec
	}
	return nil
}
