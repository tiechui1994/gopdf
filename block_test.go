package gopdf

import (
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

const (
	BLOCK_IG = "IPAexG"
	BLOCK_MC = "microsoft"
)

func BlockReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: BLOCK_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: BLOCK_MC,
		FileName: "example//ttf/microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(BlockReportExecutor), core.Detail)

	r.Execute("block_test.pdf")
	r.SaveAtomicCellText("block_test.txt")
}
func BlockReportExecutor(report *core.Report) {
	font := core.Font{Family: BLOCK_MC, Size: 13}
	block := NewTextBlock(report, font)
	block.SetContent("	啊啊飒1919290121=211-=1212==1212-01201212-012-0120-120-120飒撒加上阿萨手机卡手机卡萨按手机卡手机卡手机卡按时间按时间就卡死机卡萨按手机卡手机卡上按手机卡时间按时间就爱上按时间按时间就暗示健康按时间按时间按时间就爱上奥斯卡上空间卡萨科技拉斯阿萨斯喀上课了按时间按手机看空间拉斯奥斯卡咔咔寿加手机就卡死")
	block.GenerateAtomicCell()
}

func TestBlock(t *testing.T) {
	BlockReport()
}
