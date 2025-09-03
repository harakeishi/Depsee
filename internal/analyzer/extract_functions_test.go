package analyzer

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestExtractFunctions(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		expectedFunctions []FuncInfo
		structMap         map[string]*StructInfo
		expectedMethods   map[string][]string // 構造体名 -> メソッド名のリスト
	}{
		{
			name: "extractFunctions_SimpleFunction",
			content: `package test

func Add(a, b int) int {
	return a + b
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "Add",
					Package: "test",
					Params: []FieldInfo{
						{Name: "a", Type: "int"},
						{Name: "b", Type: "int"},
					},
					Results: []FieldInfo{
						{Name: "", Type: "int"},
					},
				},
			},
			structMap:       map[string]*StructInfo{},
			expectedMethods: map[string][]string{},
		},
		{
			name: "extractFunctions_Method",
			content: `package test

type User struct {
	Name string
}

func (u *User) GetName() string {
	return u.Name
}

func (u User) SetName(name string) {
	u.Name = name
}`,
			expectedFunctions: []FuncInfo{},
			structMap: map[string]*StructInfo{
				"User": {
					Name:    "User",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "Name", Type: "string"},
					},
					Methods: []FuncInfo{},
				},
			},
			expectedMethods: map[string][]string{
				"User": {"GetName", "SetName"},
			},
		},
		{
			name: "extractFunctions_MixedFunctionsAndMethods",
			content: `package test

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "NewCalculator",
					Package: "test",
					Params:  []FieldInfo{},
					Results: []FieldInfo{
						{Name: "", Type: "*Calculator"},
					},
				},
				{
					Name:    "Multiply",
					Package: "test",
					Params: []FieldInfo{
						{Name: "a", Type: "int"},
						{Name: "b", Type: "int"},
					},
					Results: []FieldInfo{
						{Name: "", Type: "int"},
					},
				},
			},
			structMap: map[string]*StructInfo{
				"Calculator": {
					Name:    "Calculator",
					Package: "test",
					Fields:  []FieldInfo{},
					Methods: []FuncInfo{},
				},
			},
			expectedMethods: map[string][]string{
				"Calculator": {"Add"},
			},
		},
		{
			name: "extractFunctions_NoFunctions",
			content: `package test

type User struct {
	Name string
}`,
			expectedFunctions: []FuncInfo{},
			structMap:         map[string]*StructInfo{},
			expectedMethods:   map[string][]string{},
		},
		{
			name: "extractFunctions_FunctionWithNoParams",
			content: `package test

func GetCurrentTime() time.Time {
	return time.Now()
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "GetCurrentTime",
					Package: "test",
					Params:  []FieldInfo{},
					Results: []FieldInfo{
						{Name: "", Type: "time.Time"},
					},
				},
			},
			structMap:       map[string]*StructInfo{},
			expectedMethods: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパースしてASTを作成
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse test content: %v", err)
			}

			// extractFunctions関数をテスト
			functions := extractFunctions(f, fset, "test.go", "test", tt.structMap)

			// 関数の検証
			if len(functions) != len(tt.expectedFunctions) {
				t.Errorf("extractFunctions() returned %d functions, expected %d", len(functions), len(tt.expectedFunctions))
			} else {
				for i, expected := range tt.expectedFunctions {
					if functions[i].Name != expected.Name {
						t.Errorf("functions[%d].Name = %q, expected %q", i, functions[i].Name, expected.Name)
					}
					if functions[i].Package != expected.Package {
						t.Errorf("functions[%d].Package = %q, expected %q", i, functions[i].Package, expected.Package)
					}
					// パラメータの検証
					if len(functions[i].Params) != len(expected.Params) {
						t.Errorf("functions[%d] has %d params, expected %d", i, len(functions[i].Params), len(expected.Params))
					} else {
						for j, expectedParam := range expected.Params {
							if functions[i].Params[j].Name != expectedParam.Name {
								t.Errorf("functions[%d].Params[%d].Name = %q, expected %q", i, j, functions[i].Params[j].Name, expectedParam.Name)
							}
							if functions[i].Params[j].Type != expectedParam.Type {
								t.Errorf("functions[%d].Params[%d].Type = %q, expected %q", i, j, functions[i].Params[j].Type, expectedParam.Type)
							}
						}
					}
					// 戻り値の検証
					if len(functions[i].Results) != len(expected.Results) {
						t.Errorf("functions[%d] has %d results, expected %d", i, len(functions[i].Results), len(expected.Results))
					} else {
						for j, expectedResult := range expected.Results {
							if functions[i].Results[j].Type != expectedResult.Type {
								t.Errorf("functions[%d].Results[%d].Type = %q, expected %q", i, j, functions[i].Results[j].Type, expectedResult.Type)
							}
						}
					}
				}
			}

			// メソッドの検証
			for structName, expectedMethodNames := range tt.expectedMethods {
				if s, exists := tt.structMap[structName]; exists {
					if len(s.Methods) != len(expectedMethodNames) {
						t.Errorf("struct %q has %d methods, expected %d", structName, len(s.Methods), len(expectedMethodNames))
					} else {
						for i, expectedMethodName := range expectedMethodNames {
							if s.Methods[i].Name != expectedMethodName {
								t.Errorf("struct %q method[%d].Name = %q, expected %q", structName, i, s.Methods[i].Name, expectedMethodName)
							}
							if s.Methods[i].Receiver != structName {
								t.Errorf("struct %q method[%d].Receiver = %q, expected %q", structName, i, s.Methods[i].Receiver, structName)
							}
						}
					}
				}
			}
		})
	}
}
