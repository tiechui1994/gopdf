package core

// 单位像素
type Config struct {
	startX, startY float64 // PDF页的开始坐标定位, 必须指定
	endX, endY     float64 // PDF页的结束坐标定位, 必须指定
	width, height  float64 // PDF页的宽度和高度, 必须指定

	contentWidth, contentHeight float64 // PDF页内容的宽度和高度, 计算得到
}

func (config *Config) checkConfig() {
	if config.startX < 0 || config.startY < 0 {
		panic("the pdf page start position invilid")
	}

	if config.endX < 0 || config.endY < 0 || config.endX <= config.startX || config.endY <= config.startY {
		panic("the pdf page end position invilid")
	}

	if config.width <= config.endX || config.height <= config.endY {
		panic("the pdf page width or height invilid")
	}

	// 关系验证
	if config.endX+config.startX != config.width || config.endY+config.startY != config.height {
		panic("the paf page config invilid")
	}
}

var defaultConfigs map[string]*Config // page -> config

/**************************************
A0 ~ A5 纸张像素表示
	'A0': [2383.94, 3370.39],
	'A1': [1683.78, 2383.94],
	'A2': [1190.55, 1683.78],
	'A3': [841.89, 1190.55],
	'A4': [595.28, 841.89],
	'A5': [419.53, 595.28],
***************************************/
func init() {
	defaultConfigs = make(map[string]*Config)

	defaultConfigs["A3"] = &Config{
		startX:        90.14,
		startY:        72.00,
		endX:          751.76,
		endY:          1118.55,
		width:         841.89,
		height:        1190.55,
		contentWidth:  661.62,
		contentHeight: 1046.55,
	}

	defaultConfigs["A4"] = &Config{
		startX:        90.14,
		startY:        72.00,
		endX:          505.14,
		endY:          769.89,
		width:         595.28,
		height:        841.89,
		contentWidth:  415,
		contentHeight: 697.89,
	}

	defaultConfigs["LTR"] = &Config{
		startX:        90.14,
		startY:        72.00,
		endX:          521.86,
		endY:          720,
		width:         612,
		height:        792,
		contentWidth:  431.72,
		contentHeight: 648,
	}
}

func Register(size string, config *Config) {
	if _, ok := defaultConfigs[size]; ok {
		return
	}
	config.checkConfig()
	config.contentWidth = config.endX - config.startX
	config.contentHeight = config.endY - config.startY
	defaultConfigs[size] = config
}
