package example

import (
	"fmt"
	"github.com/tiechui1994/gopdf/core"
	"testing"
)

func SimpleLink() {
	r := core.CreateReport()
	r.SetPage("A4", "P")
	r.FisrtPageNeedHeader = true
	r.FisrtPageNeedFooter = true

	r.RegisterExecutor(core.Executor(SimpleLinkExecutor), core.Detail)

	r.Execute(fmt.Sprintf("simple_link_test.pdf"))
	r.SaveAtomicCellText("simple_link_test.txt")
}

func SimpleLinkExecutor(report *core.Report) {
	var (
		lineHight = 16.0
	)
	report.SetFont(core.FontSans, 12)
	report.SetMargin(10, 20)
	x, y := report.GetXY()
	report.ExternalLink(x, y, lineHight, "外部链接(百度)", "https://www.baidu.com")
	report.InternalLinkAnchor(x, y+4*lineHight, lineHight, "内部链接", "1")

	report.AddNewPage(false)
	report.SetMargin(10, 20)
	x, y = report.GetXY()
	report.InternalLinkLink(x, y, "Hello world", "1")

}

func TestSimpleLink(t *testing.T) {
	SimpleLink()
}
