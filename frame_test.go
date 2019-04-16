package gopdf

import (
	"strings"
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

const (
	FRAME_IG = "IPAexG"
	FRAME_MD = "MPBOLD"
)

func ComplexFrameReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: FRAME_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: FRAME_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(ComplexFrameReportExecutor), core.Detail)

	r.Execute("frame_test.pdf")
	r.SaveAtomicCellText("frame_test.txt")
}

func ComplexFrameReportExecutor(report *core.Report) {
	unit := report.GetUnit()
	font := core.Font{Family: FRAME_MD, Size: 10}

	report.Font(FRAME_MD, 10, "")
	report.SetFont(FRAME_MD, 10)

	lineSpace := 0.1 * unit
	lineHeight := report.MeasureTextWidth("中") / unit

	frame := NewFrameWithWidth(48*unit, lineHeight, lineSpace, report)
	frame.SetFrameType(FRAME_DOTTED)
	frame.SetBackColor("222,111,11")
	frame.SetFont(font)
	frame.SetMarign(core.NewScope(20, 50, 0, 0))
	frame.SetBorder(core.NewScope(4, 50, 0, 0))

	content := `13.2.10 Subquery  Syntax 
13.2.10.1 The Subquery as Scalar Operand 
13.2.10.2 Comparisons Using Subqueries
13.2.10.3 Subqueries with ANY, IN, or SOME
13.2.10.4 Subqueries with ALL
13.2.10.5 Row Subqueries
13.2.10.6 Subqueries with EXISTS or NOT EXISTS
13.2.10.7 Correlated Subqueries
13.2.10.8 Derived Tables
13.2.10.9 Subquery Errors
13.2.10.10 Optimizing Subqueries
13.2.10.11 Rewriting Subqueries as Joins
A subquery is a SELECT statement within another statement.
All subquery forms and operations that the SQL standard requires are supported, as well as a few features that are MySQL-specific.
Here is an example of a subquery:
SELECT * FROM t1 WHERE column1 = (SELECT column1 FROM t2);
In this example, SELECT * FROM t1 ... is the outer query (or outer statement), and (SELECT column1 FROM t2) is the subquery. We say that the subquery is nested within the outer query, and in fact it is possible to nest subqueries within other subqueries, to a considerable depth. A subquery must always appear within parentheses.
The main advantages of subqueries are:
They allow queries that are structured so that it is possible to isolate each part of a statement.
They provide alternative ways to perform operations that would otherwise require complex joins and unions.
Many people find subqueries more readable than complex joins or unions. Indeed, it was the innovation of subqueries that gave people the original idea of calling the early SQL “Structured Query Language.”
how the optimizer handles subqueries, see Section 8.2.2, “Optimizing Subqueries, Derived Tables, and View References”. For a discussion of restrictions on subquery use, including performance issues for certain forms of subquery syntax, see Section C.4, “Restrictions on Subqueries”.`
	frame.SetContent(strings.Repeat(content, 4))
	frame.GenerateAtomicCell()
}

func TestComplexFrameReport(t *testing.T) {
	ComplexFrameReport()
}
