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
			name: "SetFilters_SingleTargetPackage",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{TargetPackages: []string{"test"}}},
		},
		{
			name: "SetFilters_MultipleTargetPackages",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{TargetPackages: []string{"test1", "test2", "test3"}}},
		},
		{
			name: "SetFilters_ExcludePackages",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"test"},
				ExcludePackages: []string{"exclude1", "exclude2"},
			}},
		},
		{
			name: "SetFilters_ExcludeDirs",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages: []string{"test"},
				ExcludeDirs:    []string{"/tmp", "/test"},
			}},
		},
		{
			name: "SetFilters_AllFilters",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"main", "utils", "config"},
				ExcludePackages: []string{"test", "mock"},
				ExcludeDirs:     []string{"/vendor", "/node_modules", "/tmp"},
			}},
		},
		{
			name: "SetFilters_EmptyFilters",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"existing"}},
			},
			args: args{filters: Filters{}},
		},
		{
			name: "SetFilters_OverwriteExistingFilters",
			fields: fields{
				Filters: Filters{
					TargetPackages:  []string{"old1", "old2"},
					ExcludePackages: []string{"oldExclude"},
					ExcludeDirs:     []string{"/oldDir"},
				},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"new1", "new2"},
				ExcludePackages: []string{"newExclude"},
				ExcludeDirs:     []string{"/newDir"},
			}},
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

			// TargetPackagesの確認
			if !slices.Equal(ga.Filters.TargetPackages, tt.args.filters.TargetPackages) {
				t.Errorf("Filters.TargetPackages = %v, want %v", ga.Filters.TargetPackages, tt.args.filters.TargetPackages)
			}

			// ExcludePackagesの確認
			if !slices.Equal(ga.Filters.ExcludePackages, tt.args.filters.ExcludePackages) {
				t.Errorf("Filters.ExcludePackages = %v, want %v", ga.Filters.ExcludePackages, tt.args.filters.ExcludePackages)
			}

			// ExcludeDirsの確認
			if !slices.Equal(ga.Filters.ExcludeDirs, tt.args.filters.ExcludeDirs) {
				t.Errorf("Filters.ExcludeDirs = %v, want %v", ga.Filters.ExcludeDirs, tt.args.filters.ExcludeDirs)
			}

		})
	}
}
