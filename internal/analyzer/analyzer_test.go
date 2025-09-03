package analyzer

import (
	"os"
	"slices"
	"testing"

	"github.com/harakeishi/depsee/internal/logger"
)

func TestMain(m *testing.M) {
	// テスト用のログ設定
	logger.Init(logger.Config{
		Level:  logger.LevelError, // テスト時はエラーのみ
		Format: "text",
		Output: os.Stderr,
	})

	code := m.Run()
	os.Exit(code)
}

func TestGoAnalyzer_SetFilters(t *testing.T) {
	type fields struct {
		Filters   Filters
		filesPath []string
		Result    *Result
	}
	type args struct {
		filters Filters
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetFilters",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{TargetPackages: []string{"test"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ga := &GoAnalyzer{
				Filters:   tt.fields.Filters,
				filesPath: tt.fields.filesPath,
				Result:    tt.fields.Result,
			}
			ga.SetFilters(tt.args.filters)
			// フィルターが設定されていることを確認
			if ga.Filters.TargetPackages == nil {
				t.Errorf("Filters.TargetPackages = nil, want %v", tt.args.filters.TargetPackages)
			}
			if !slices.Equal(ga.Filters.TargetPackages, tt.args.filters.TargetPackages) {
				t.Errorf("Filters.TargetPackages = %v, want %v", ga.Filters.TargetPackages, tt.args.filters.TargetPackages)
			}
		})
	}
}
