package example

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/tiechui1994/gopdf"
	"github.com/tiechui1994/gopdf/core"
)

type Template struct {
	Modelname       string      `json:"modelname"`
	MultiSelectList []string    `json:"multiSelectList"`
	Value           interface{} `json:"value"`
	ValueList       interface{} `json:"valueList"`
	Modeltype       string      `json:"modeltype"`
	NeedPrint       string      `json:"needPrint"`

	TableObj struct {
		Columns []struct {
			ColumnsName string `json:"columnsName"`
		} `json:"columns"` // 列名称
		Rows []string `json:"rows"` // 行名称, 从第2行开算, 第一行是Header,没有
	} `json:"tableObj"`

	TbodyArray [][]struct {
		ColumnsType    string      `json:"columnsType"` // 当前TableCell类型
		ColumnsName    string      `json:"columnsName"` // 当前TableCell对应列的名称
		RowName        string      `json:"rowsName"`    // 当前TableCell对于的行名称
		Content        interface{} `json:"content"`     // 内容
		SelectedOption string      `json:"selectedOption"`
	} `json:"tbodyArray"`

	TableTotalArr []interface{} `json:"tableTotalArr"` // 合计
}

func handleTable(template Template) (rows, cols int, cells [][]string, hasRowName bool) {
	var (
		colNames = template.TableObj.Columns
		rowNames = template.TableObj.Rows
		body     = template.TbodyArray
	)

	hasRowName = len(rowNames) != 0
	cols = len(colNames)
	rows = 1 // 头

	if hasRowName {
		cols += 1             // 第0行
		rows += len(rowNames) // 其余的的行
	} else {
		if isEmpty(body) {
			rows += 1 // 至少有一行
		} else {
			rows += len(body)
		}
	}

	if !isEmpty(template.TableTotalArr) {
		cells = make([][]string, rows+1)
	} else {
		cells = make([][]string, rows)
	}
	for i := range cells {
		cells[i] = make([]string, cols)
	}

	// header
	if hasRowName {
		cells[0][0] = ""
	}

	for i := range colNames {
		if hasRowName {
			cells[0][i+1] = colNames[i].ColumnsName
		} else {
			cells[0][i] = colNames[i].ColumnsName
		}
	}

	// 拥有rowname, 肯定有统计信息
	if hasRowName {
		// body内容: rows - 2 * cols - 1
		for i := 1; i < rows; i++ {
			cells[i][0] = body[i-1][0].RowName
			for j := 0; j < cols-1; j++ {
				if body[i-1][j].ColumnsType == "select" {
					cells[i][j+1] = body[i-1][j].SelectedOption
					continue
				}

				switch body[i-1][j].Content.(type) {
				case string:
					cells[i][j+1] = body[i-1][j].Content.(string)
				case float64:
					cells[i][j+1] = fmt.Sprintf("%0.2f", body[i-1][j].Content)
				}
			}
		}

		// 最后一行
		if !isEmpty(template.TableTotalArr) {
			cells[rows][0] = "合计"
			for j := 0; j < cols-1; j++ {
				switch template.TableTotalArr[j].(type) {
				case string:
					cells[rows][j+1] = template.TableTotalArr[j].(string)
				case float64:
					cells[rows][j+1] = fmt.Sprintf("%.2f", template.TableTotalArr[j].(float64))
				}
			}

			return rows + 1, cols, cells, hasRowName
		}

		return rows, cols, cells, hasRowName
	}

	// 没有rowname, 但是有内容
	if !hasRowName && !isEmpty(body) {
		for i := 1; i < rows; i++ {
			for j := 0; j < cols; j++ {
				if body[i-1][j].ColumnsType == "select" {
					cells[i][j] = body[i-1][j].SelectedOption
					continue
				}

				switch body[i-1][j].Content.(type) {
				case string:
					cells[i][j] = body[i-1][j].Content.(string)
				case float64:
					cells[i][j] = fmt.Sprintf("%0.2f", body[i-1][j].Content.(float64))
				}
			}
		}

		// 合计
		if !isEmpty(template.TableTotalArr) {
			for j := 0; j < cols; j++ {
				switch template.TableTotalArr[j].(type) {
				case string:
					cells[rows][j] = template.TableTotalArr[j].(string)
				case float64:
					cells[rows][j] = fmt.Sprintf("%.2f", template.TableTotalArr[j].(float64))
				}
			}

			return rows + 1, cols, cells, hasRowName
		}

		return rows, cols, cells, hasRowName
	} else {
		for j := 0; j < cols; j++ {
			cells[1][j] = ""
		}
	}

	return rows, cols, cells, hasRowName
}
func isEmpty(object interface{}) bool {
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}
func handleTemplateRow(template Template) (isTable bool, key, value string) {
	key = template.Modelname

	switch template.Modeltype {
	case "1", "2", "5", "9":
		if template.Value == nil {
			template.Value = ""
		}
		value = template.Value.(string)

	case "4":
		if template.Value == nil {
			template.Value = ""
		}
		var temp = &[]struct {
			SourceID    string `json:"source_id"`
			FileName    string `json:"file_name"`
			DownloadUrl string `json:"downloadUrl"`
		}{}
		err := json.Unmarshal([]byte(template.Value.(string)), temp)
		if err != nil {
			return false, key, ""
		}
		if isEmpty(temp) {
			value = ""
			return
		}

		for _, v := range *temp {
			value += fmt.Sprintf("  %s", v.FileName)
		}
	case "6":
		m, ok := template.ValueList.([]interface{})
		if ok == false {
			value = template.ValueList.(string)
		} else {
			var values string
			for _, v := range m {
				values += fmt.Sprintf("  %s", v)
			}
			value = values
		}

	case "7":
		key = ""

	case "8":
	}

	return template.Modeltype == "8", key, value
}

func SimpleTableReportExecutor(report *core.Report) {
	var (
		templates []Template
		unit      = report.GetUnit()
		lineSpace = 0.01 * unit
		lineHight = 1.9 * unit
	)
	str := `[
	{"modelname":"所在公司","modeltype":"5","value":"中国航天科技九院"},
	{"modelname":"报销部门","modeltype":"1","value":"财务部"},
	{"modelname":"报销人","modeltype":"1","value":"钱伟长"},
	{"modelname":"职务","modeltype":"1","value":"技术总领导"},
	{"modelname":"项目名称","modeltype":"1","value":"航天科技人员培训"},
	{"modelname":"出差时间","modeltype":"2","value":"12.21"},
	{"modelname":"结束时间","modeltype":"2","value":"12.27"},
	{"modelname":"目的地","modeltype":"1","value":"青海金银滩"},
	{"modelname":"出差说明","modeltype":"2","value":"指导培训"},
	{"modelname":"实际出差时间和结束时间段","modeltype":"2"},
	{
		"modelname":"报销明细单",
		"modeltype":"8",
		"tableObj":{
			"columns":[
				{"columnsName":"出发时间","columnsType":"text"},
				{"columnsName":"出发地点","columnsType":"text"},
				{"columnsName":"到达时间","columnsType":"text"},
				{"columnsName":"到达地点","columnsType":"text"},
				{"columnsName":"交通工具","columnsType":"text"},
				{"columnsName":"事项","columnsType":"text"},
				{"columnsName":"交通费","columnsType":"number"},
				{"columnsName":"住宿费","columnsType":"number"},
				{"columnsName":"招待费","columnsType":"number"},
				{"columnsName":"补贴","columnsType":"number"},
				{"columnsName":"票据数","columnsType":"text"},
				{"columnsName":"票据序号","columnsType":"text"},
				{"columnsName":"备注","columnsType":"text"}
			],
			"rows":[]
		},
		"tbodyArray":[],
		"tableTotalArr":["","","","","","","","","","","","",""]
	},
	{"modelname":"报销总额（小写）", "modeltype":"1","value":""},
	{"modelname":"报销总额（大写）","modeltype":"1","tableObj":{"columns":[],"rows":[]}},
	{"modelname":"超支金额","modeltype":"1","value":""},
	{"modelname":"应报金额","modeltype":"1","value":""},
	{
		"modelname":"调整金额",
		"modeltype":"8",
		"tableObj":{
			"columns":[
				{"columnsName":"交通费","columnsType":"text"},
				{"columnsName":"住宿费","columnsType":"text"},
				{"columnsName":"招待费","columnsType":"text"},
				{"columnsName":"补贴","columnsType":"text"}
			],
			"rows":["调整","实报金额"]
		},
		"tbodyArray":[
			[
				{"columnsName":"交通费","rowsName":"调整"},
				{"columnsName":"住宿费","rowsName":"调整"},
				{"columnsName":"招待费","rowsName":"调整"},
				{"columnsName":"补贴","rowsName":"调整"}
			],
			[
				{"columnsName":"交通费","rowsName":"实报金额"},
				{"columnsName":"住宿费","rowsName":"实报金额"},
				{"columnsName":"招待费","rowsName":"实报金额"},
				{"columnsName":"补贴","rowsName":"实报金额"}
			]
		],
		"tableTotalArr":["","","",""]
	},
	{"modelname":"备注","modeltype":"2", "value":""},
	{"modelname":"附件","modeltype":"4", "fileNameArr":[],"attachfileArr":[],"value":"[]"},
	{"modelname":"附单据","modeltype":"4","fileNameArr":[],"attachfileArr":[]}
]`

	json.Unmarshal([]byte(str), &templates)

	for _, template := range templates {
		isTable, key, value := handleTemplateRow(template)
		// key != "" 是过滤 Modeltype 为 "7"的情况
		if !isTable && key != "" {
			report.SetMargin(4*unit, 0)
			content := fmt.Sprintf("%s: %s", key, value)
			contentDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
			contentDiv.SetFont(textFont).SetContent(content).GenerateAtomicCell()
			report.SetMargin(0, 1*unit)

		}

		// 处理表格
		if isTable {
			report.SetMargin(4*unit, 0)
			content := fmt.Sprintf("%s:", key)
			contentDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
			contentDiv.SetFont(textFont).SetContent(content).GenerateAtomicCell()
			report.SetMargin(0, 0.5*unit)

			rows, cols, cells, hasRowName := handleTable(template)
			table := gopdf.NewTable(cols, rows, 100*unit, lineHight, report)
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					cell := table.NewCell()
					element := gopdf.NewTextCell(table.GetColWidth(i, j), lineHight, lineSpace, report)
					element.SetFont(textFont)
					element.SetBorder(core.Scope{Left: 0.5 * unit, Top: 0.5 * unit})
					if i == 0 || j == 0 && hasRowName {
						element.HorizontalCentered()
					}
					element.SetContent(cells[i][j])
					cell.SetElement(element)
				}
			}

			table.GenerateAtomicCell()
			report.SetMargin(0, 1*unit)
		}
	}
}

func SimpleTableReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: FONT_MY,
		FileName: "ttf//microsoft.ttf",
	}
	font2 := core.FontMap{
		FontName: FONT_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(SimpleTableReportExecutor), core.Detail)
	r.Execute("simple_table_test.pdf")
	r.SaveAtomicCellText("simple_table_test.txt")
}

func TestComplexTableReport(t *testing.T) {
	SimpleTableReport()
}
