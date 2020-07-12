package core

import "fmt"

// the units of below is pixels.
type Config struct {
	startX, startY float64 // PDF page start position
	endX, endY     float64 // PDF page end postion
	width, height  float64 // PDF page width and height

	contentWidth, contentHeight float64 // PDF page content width and height
}

// Params width, height is pdf page width and height
// Params padingH, padingV is pdf horizontal and vertical padding
// The units of the above parameters are pixels.
// Params width must more than 2*padingH, and height must more 2*padingV
func NewConfig(width, height float64, padingH, padingV float64) (*Config, error) {
	if width <= 0 || height <= 0 || padingH < 0 || padingV < 0 {
		return nil, fmt.Errorf("params must more than zero")
	}

	if width <= 2*padingH || height <= 2*padingV {
		return nil, fmt.Errorf("this config params invalid")
	}

	c := &Config{
		width:  width,
		height: height,

		startX: padingH,
		startY: padingV,

		contentWidth:  width - 2*padingH,
		contentHeight: height - 2*padingV,
	}

	c.endX = c.startX + c.contentWidth
	c.endY = c.startY + c.contentHeight

	return c, nil
}

func (config *Config) GetWidthAndHeight() (width, height float64) {
	return config.width, config.width
}

// Get pdf page start position, from the position you can write the pdf body content.
func (config *Config) GetStart() (x, y float64) {
	return config.startX, config.startY
}

func (config *Config) GetEnd() (x, y float64) {
	return config.endX, config.endY
}

var defaultConfigs map[string]*Config // page -> config

/**************************************
A0 ~ A5 page width and height config:
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

// Register create self pdf config
func Register(size string, config *Config) {
	if _, ok := defaultConfigs[size]; ok {
		panic("config size has exist")
	}

	defaultConfigs[size] = config
}
