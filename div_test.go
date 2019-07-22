package gopdf

import (
	"testing"
	"strings"

	"github.com/tiechui1994/gopdf/core"
)

const (
	DIV_IG = "IPAexG"
	DIV_MD = "MPBOLD"
)

func DivReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: DIV_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: DIV_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(DivReportExecutor), core.Detail)

	r.Execute("div_test.pdf")
	r.SaveAtomicCellText("div_test.txt")
}
func DivReportExecutor(report *core.Report) {
	font := core.Font{Family: DIV_MD, Size: 10}

	report.Font(DIV_MD, 10, "")
	report.SetFont(DIV_MD, 10)

	lineSpace := 1.0
	lineHeight := report.MeasureTextWidth("中")

	div := NewDivWithWidth(300, lineHeight, lineSpace, report)
	div.SetFont(font)
	div.SetBackColor("11,123,244")
	div.RightAlign()
	div.SetMarign(core.NewScope(10, 20, 0, 0))
	div.SetBorder(core.NewScope(10, 50, 0, 0))
	div.SetContent(`13.2.10 Subquery  Syntax 
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

Here is an example statement that shows the major points about subquery syntax as specified by the SQL standard and supported in MySQL:

DELETE FROM t1
WHERE s11 > ANY
 (SELECT COUNT(*) /* no hint */ FROM t2
  WHERE NOT EXISTS
   (SELECT * FROM t3
    WHERE ROW(5*t2.s1,77)=
     (SELECT 50,11*s1 FROM t4 UNION SELECT 50,77 FROM
      (SELECT * FROM t5) AS t5)));
A subquery can return a scalar (a single value), a single row, a single column, or a table (one or more rows of one or more columns). These are called scalar, column, row, and table subqueries. Subqueries that return a particular kind of result often can be used only in certain contexts, as described in the following sections.

There are few restrictions on the type of statements in which subqueries can be used. A subquery can contain many of the keywords or clauses that an ordinary SELECT can contain: DISTINCT, GROUP BY, ORDER BY, LIMIT, joins, index hints, UNION constructs, comments, functions, and so on.

A subquery's outer statement can be any one of: SELECT, INSERT, UPDATE, DELETE, SET, or DO.

In MySQL, you cannot modify a table and select from the same table in a subquery. This applies to statements such as DELETE, INSERT, REPLACE, UPDATE, and (because subqueries can be used in the SET clause) LOAD DATA INFILE.

For information about how the optimizer handles subqueries, see Section 8.2.2, “Optimizing Subqueries, Derived Tables, and View References”. For a discussion of restrictions on subquery use, including performance issues for certain forms of subquery syntax, see Section C.4, “Restrictions on Subqueries”.`)
	div.GenerateAtomicCell()
}

func ComplexDivReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: DIV_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: DIV_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ComplexDivReportExecutor), core.Detail)

	r.Execute("complex_div_test.pdf")
	r.SaveAtomicCellText("complex_div_test.txt")
}
func ComplexDivReportExecutor(report *core.Report) {
	font := core.Font{Family: DIV_MD, Size: 10}

	report.Font(DIV_MD, 10, "")
	report.SetFont(DIV_MD, 10)

	lineSpace := 1.0
	lineHeight := report.MeasureTextWidth("中")

	frame := NewDivWithWidth(415, lineHeight, lineSpace, report)
	frame.SetFrameType(DIV_DOTTED)
	//frame.SetBackColor("222,111,11")
	frame.SetFont(font)
	frame.SetMarign(core.NewScope(120, 50, 0, 0))
	frame.SetBorder(core.NewScope(10, 50, 10, 10))

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

func TestDivReport(t *testing.T) {
	DivReport()
}

func TestComplexDivReport(t *testing.T) {
	ComplexDivReport()
}
