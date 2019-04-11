package core

// 当作为Margin的时候, Right无法生效
// 当作为Border的时候, Bottom无法生效
type Scope struct {
	Left   float64
	Top    float64
	Right  float64
	Bottom float64
}

func NewScope(left, top, right, bottom float64) Scope {
	return Scope{Left: left, Top: top, Right: right, Bottom: bottom}
}

func (scope *Scope) ReplaceBorder() {
	if scope.Left < 0 {
		scope.Left = 0
	}
	if scope.Right < 0 {
		scope.Right = 0
	}
	if scope.Top < 0 {
		scope.Top = 0
	}

	scope.Bottom = 0
}

func (scope *Scope) ReplaceMarign() {
	scope.Right = 0
	if scope.Bottom < 0 {
		scope.Bottom = 0
	}
}

type Font struct {
	Family string // 字体名称
	Style  string // 字体风格, 目前支持, "" , "U", "B","I", 其中"B", "I" 需要字体本身定义
	Size   int    // 字体大小
}

type Cell interface {
	GenerateAtomicCell(height float64) (writed, remain int, err error) // 写入的行数, 剩余的行数,错误
	TryGenerateAtomicCell(height float64) (writed, remain int)         // 尝试写入
	GetHeight() (height float64)                                       // 当前cell的height
	GetLastHeight() (height float64)                                   // 最近一次cell的height
}
