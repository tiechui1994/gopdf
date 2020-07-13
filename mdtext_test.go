package gopdf

import (
	"testing"

	"github.com/tiechui1994/gopdf/core"
	"encoding/base64"
)

const (
	MD_IG = "IPAexG"
	MD_MC = "Microsoft"
	MD_MB = "Microsoft Bold"
)

func MdReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: MD_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: MD_MC,
		FileName: "example//ttf/microsoft.ttf",
	}
	font3 := core.FontMap{
		FontName: MD_MB,
		FileName: "example//ttf/microsoft-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(MdReportExecutor), core.Detail)

	r.Execute("block_test.pdf")
	r.SaveAtomicCellText("block_test.txt")
}

func MdReportExecutor(report *core.Report) {
	text, _ := base64.StdEncoding.DecodeString("PiDms6jmhI86IAo+Cj4gMS7lnKjpk77mjqVD5bqT55qE5L2/55SoLCDkuI3mlK/mjIHmnaHku7bpgInmi6kuIOW5tuS4lENHT+WPguaVsOacieS4peagvOeahOagvOW8jyBgI2NnbyBDRkxBR1M6Li4uYCDmiJbogIUgCj4gYCNjZ28gTERGTEFHUzogLi4uIGAsIOWNsyBgI2Nnb2Ag5ZKMIOWPguaVsChgQ0ZMQUdTYCwgYExERkxBR1NgKSAKPiAKPiAyLuWvueS6jkPor63oqIDlupMoYC5oYCDmlofku7blrprkuYnlhoXlrrkg5ZKMIGAuY2Ag5paH5Lu25a6e546wIGAuaGAg55qE5a6a5LmJKSwg5ZyoQ0dP5b2T5Lit5byV55SoIGAuaGAg5paH5Lu2LCDlv4Xpobvph4fnlKgKPiBg5Yqo5oCB5bqTL+mdmeaAgeW6k2Ag6ZO+5o6l55qE5pa55byPLCDlkKbliJnlj6/og73ml6Dms5XnvJbor5HpgJrov4cuICA=")

	mt, _ := NewMarkdownText(report, 10, map[string]string{
		FONT_NORMAL: MD_MC,
		FONT_IALIC:  MD_MC,
		FONT_BOLD:   MD_MB,
	})
	mt.SetText(string(text))

	ctetx := content{
		pdf:  report,
		Type: TEXT_NORMAL,
	}
	ctetx.SetText(MD_MC, "中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888")

	btetx := content{
		pdf:  report,
		Type: TEXT_WARP,
	}
	btetx.SetText(MD_MC, "中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888")
	itetx := content{
		pdf:  report,
		Type: TEXT_NORMAL,
	}
	ctetx.GenerateAtomicCell()
	itetx.SetText(MD_MB, "中田田日日000088888空间卡生萨斯卡拉斯科拉11209901接口就爱看手机爱卡就爱29912-0-接口就爱看手机爱卡就爱021-012-12012-021-012-0-021-0120-1201221")
	btetx.GenerateAtomicCell()
	itetx.GenerateAtomicCell()
}

func TestMd(t *testing.T) {
	MdReport()
}

func TestContent(t *testing.T) {
}
