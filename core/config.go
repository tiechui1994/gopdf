package core

// 单位像素
type Config struct {
	startX, startY float64 // PDF页的开始坐标定位, 必须指定
	endX, endY     float64 // PDF页的结束坐标定位, 必须指定
	width, height  float64 // PDF页的宽度和高度, 必须指定

	contentWidth, contentHeight float64 // PDF页内容的宽度和高度, 计算得到
}

func (c *Config) checkConfig() {
	if c.startX < 0 || c.startY < 0 {
		panic("the pdf page start position invilid")
	}

	if c.endX < 0 || c.endY < 0 || c.endX <= c.startX || c.endY <= c.startY {
		panic("the pdf page end position invilid")
	}

	if c.width <= c.endX || c.height <= c.endY {
		panic("the pdf page width or height invilid")
	}

	// 关系验证
	if c.endX+c.startX != c.width || c.endY+c.startY != c.height {
		panic("the paf page config invilid")
	}
}

var defaultConfigs map[string]Config // page -> config

func init() {
	defaultConfigs = make(map[string]Config)

	defaultConfigs["A4"] = Config{
		startX:        90.14,
		startY:        72.00,
		endX:          505.14,
		endY:          769.89,
		width:         595.28,
		height:        841.89,
		contentWidth:  415,
		contentHeight: 697.89,
	}

	defaultConfigs["LTR"] = Config{
		startX:        90.14,
		startY:        72.00,
		endX:          505.14,
		endY:          769.89,
		width:         612,
		height:        792,
		contentWidth:  415,
		contentHeight: 697.89,
	}
}

func Register(size string, config Config) {
	if _, ok := defaultConfigs[size]; ok {
		return
	}
	config.checkConfig()
	config.contentWidth = config.endX - config.startX
	config.contentHeight = config.endY - config.startY
	defaultConfigs[size] = config
}
