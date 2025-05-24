package reserved

// 予約語を含む構造体とインターフェースのテスト

// graph は mermaid の予約語
type graph struct {
	ID   int
	Name string
}

// end も mermaid の予約語
type end struct {
	Value string
}

// defaultConfig は mermaid の予約語 default を含む
type defaultConfig struct {
	Config string
}

// circle も mermaid の予約語
type circle struct {
	Radius float64
}

// 通常の構造体
type User struct {
	Graph   *graph
	End     *end
	Default *defaultConfig
	Circle  *circle
}

// 予約語を含む関数
func CreateGraph() *graph {
	return &graph{}
}

func ProcessEnd(e *end) {
	// 処理
}

func SetDefault(d *defaultConfig) {
	// 処理
}

// 予約語を含むインターフェース
type node interface {
	GetID() int
}

type link interface {
	Connect(from, to node)
}
