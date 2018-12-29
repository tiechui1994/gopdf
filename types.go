package gopdf

type Scope struct {
	Left   float64
	Top    float64
	Right  float64
	Bottom float64
}

type Font struct {
	Family string // 字体名称
	Style  string // 字体风格, 目前支持, "" , "U", "B","I", 其中"B", "I" 需要字体本身定义
	Size   int    // 字体大小
}

type Element interface {
	GenerateAtomicCell() error
	GetHeight() float64
	SetHeight(height float64)
	SetBorder(border Scope)
}
