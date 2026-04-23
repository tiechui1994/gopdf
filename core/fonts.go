package core

// Logical font names for embedded Alibaba PuHuiTi. Layout uses PDF points (pt) only;
// text uses registered TTF — not PDF built-in fonts.
const (
	FontSans     = "gopdf-sans"
	FontSansBold = "gopdf-sans-bold"
)

// DefaultFontMaps returns embedded Alibaba Regular and Bold (no filesystem paths).
func DefaultFontMaps() []*FontMap {
	reg := make([]byte, len(embeddedAlibabaSans))
	copy(reg, embeddedAlibabaSans)
	bold := make([]byte, len(embeddedAlibabaSansBold))
	copy(bold, embeddedAlibabaSansBold)
	return []*FontMap{
		{FontName: FontSans, Data: reg, FileName: "AlibabaPuHuiTi-Regular.ttf"},
		{FontName: FontSansBold, Data: bold, FileName: "AlibabaPuHuiTi-Bold.ttf"},
	}
}

func cloneFontMaps(src []*FontMap) []*FontMap {
	if len(src) == 0 {
		return nil
	}
	out := make([]*FontMap, len(src))
	for i := range src {
		c := *src[i]
		if len(src[i].Data) > 0 {
			c.Data = append([]byte(nil), src[i].Data...)
		}
		out[i] = &c
	}
	return out
}
