package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
	"testing"
)

func ComplexHLineReport() {
	r := core.CreateReport()
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(ComplexHLineReportExecutor), core.Detail)

	r.Execute("hr_test.pdf")
	r.SaveAtomicCellText("hr_test.txt")
}
func ComplexHLineReportExecutor(report *core.Report) {
	hr := NewHLine(report)
	hr.SetColor(0)
	hr.GenerateAtomicCell()
}

func TestComplexHLineReport(t *testing.T) {
	ComplexHLineReport()
}
