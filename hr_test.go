package gopdf

import (
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

func ComplexHLineReport() {
	r := core.CreateReport()
	if err := r.SetPage("A4", "P"); err != nil {
		panic(err)
	}

	r.RegisterExecutor(core.Executor(ComplexHLineReportExecutor), core.Detail)

	if err := r.Execute("hr_test.pdf"); err != nil {
		panic(err)
	}
	if err := r.SaveAtomicCellText("hr_test.txt"); err != nil {
		panic(err)
	}
}
func ComplexHLineReportExecutor(report *core.Report) {
	// Line width in pt (≈5 mm).
	const widthPt = 5 * 2.834645669
	hr := NewHLine(report)
	hr.SetColor(0.7).SetWidth(widthPt)
	hr.GenerateAtomicCell()
}

func TestComplexHLineReport(t *testing.T) {
	ComplexHLineReport()
}

type A interface {
	A()
}

type B interface {
	B()
}

type T struct {
}

func (t *T) A() {

}
func (t *T) B() {

}

func TestAB(t *testing.T) {
	var b B
	b = &T{}
	if _, ok := b.(A); ok {
		t.Log("ok")
	}
}
