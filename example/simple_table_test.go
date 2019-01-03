package example

import (
	"fmt"
	"testing"
	"reflect"
	"encoding/json"

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
		} `json:"columns"`          // 列名称
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
		rows += 1             // 合计必须有
		rows += len(rowNames) // 其余的的行
	} else {
		if isEmpty(body) {
			rows += 1 // 至少有一行
		} else {
			rows += len(body)
		}
	}

	cells = make([][]string, rows)
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
		for i := 1; i < rows-1; i++ {
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
		cells[rows-1][0] = "合计"
		for j := 0; j < cols-1; j++ {
			switch template.TableTotalArr[j].(type) {
			case string:
				cells[rows-1][j+1] = template.TableTotalArr[j].(string)
			case float64:
				cells[rows-1][j+1] = fmt.Sprintf("%.2f", template.TableTotalArr[j])
			}
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
	//str := `[{"modelname":"时间","modeldesc":"耗费时长","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"1","enumList":[],"multiSelectList":[],"tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13673","value":"11","isEdit":true,"canShow":true},{"modelname":"测试","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"4","enumList":[],"multiSelectList":[],"modeldesc":"任务1","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13674","fileNameArr":["3.png"],"attachfileArr":[{"source_id":"f0ec41b0b39ee40b1b60cbd949accd72accc36b5ab078365ae090724000d0815","file_name":"3.png"}],"value":"[{\"source_id\":\"f0ec41b0b39ee40b1b60cbd949accd72accc36b5ab078365ae090724000d0815\",\"file_name\":\"3.png\",\"downloadUrl\":\"/cloudproject/v1/query/file?sourceid=f0ec41b0b39ee40b1b60cbd949accd72accc36b5ab078365ae090724000d0815&loginSession=1e0d3d54caf2001ca3c1531812d1febb&requserid=000fc2937c7d706845561141fc68ef91\",\"isImg\":true}]","multi_attach_buff":[{"source_id":"f0ec41b0b39ee40b1b60cbd949accd72accc36b5ab078365ae090724000d0815","file_name":"3.png","downloadUrl":"/cloudproject/v1/query/file?sourceid=f0ec41b0b39ee40b1b60cbd949accd72accc36b5ab078365ae090724000d0815&loginSession=1e0d3d54caf2001ca3c1531812d1febb&requserid=000fc2937c7d706845561141fc68ef91","isImg":true}],"isEdit":true,"canShow":true},{"modelname":"测试1","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"123","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13675","value":"1.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n2.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n3.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n4.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n1.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n2.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n3.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择\n4.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择","isEdit":true,"canShow":true},{"modelname":"测试2","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"5","enumList":["111","2222"],"multiSelectList":[],"modeldesc":"1234","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13676","value":"111","isEdit":true,"canShow":true},{"modelname":"测试3","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"6","enumList":[],"multiSelectList":["123","345","678"],"modeldesc":"333","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13677","value":{"0":true,"1":true,"2":true,"$$hashKey":"object:13690"},"valueList":["123","345","678"],"isEdit":true,"canShow":true},{"modelname":"测试4","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"7","enumList":[],"multiSelectList":[],"modeldesc":"233","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:13678","isEdit":true,"value":"","canShow":true},{"modelname":"测试5","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"8","enumList":[],"multiSelectList":[],"modeldesc":"555","tableObj":{"columns":[{"columnsName":"文本","columnsType":"text","$$hashKey":"object:13273"},{"columnsName":"数字无总和","columnsType":"number_noSum","$$hashKey":"object:13274"},{"columnsName":"数字+总和","columnsType":"number","$$hashKey":"object:13275"},{"columnsName":"下拉框","columnsType":"select","selectList":["234","333","33"],"$$hashKey":"object:13276"}],"rows":["123","456"]},"$$hashKey":"object:13679","tbodyArray":[[{"columnsName":"文本","rowsName":"123","columnsType":"text","content":"1.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 2.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 3.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 4.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择"},{"columnsName":"数字无总和","rowsName":"123","columnsType":"number_noSum","content":96},{"columnsName":"数字+总和","rowsName":"123","columnsType":"number","content":55},{"columnsName":"下拉框","rowsName":"123","columnsType":"select","content":"","selectedOption":"234","selectList":["234","333","33"]}],[{"columnsName":"文本","rowsName":"456","columnsType":"text","content":"1.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 2.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 3.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择 4.后台模板节点添加默认负责人时，下拉框选择优化，当选择有值时，也可以进行选择"},{"columnsName":"数字无总和","rowsName":"456","columnsType":"number_noSum","content":85},{"columnsName":"数字+总和","rowsName":"456","columnsType":"number","content":33},{"columnsName":"下拉框","rowsName":"456","columnsType":"select","content":"","selectedOption":"234","selectList":["234","333","33"]}]],"tableTotalArr":["","",88,""],"hasNum":true,"isEdit":true,"value":"","canShow":true},{"modelname":"测试6","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"9","enumList":[],"multiSelectList":[],"modeldesc":"时间","tableObj":{"columns":[],"rows":[]},"dateForm":"1","$$hashKey":"object:22870","value":"2019-01-03","isEdit":true,"canShow":true},{"modelname":"测试7","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"9","enumList":[],"multiSelectList":[],"modeldesc":"时间加日期","tableObj":{"columns":[],"rows":[]},"dateForm":"2","$$hashKey":"object:22881","value":"2019-01-17 15:50","isEdit":true,"canShow":true}]`
	str := `[{"modelname":"所在公司","isRequired":"1","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"5","enumList":["杭州古北电子科技有限公司","杭州古北进出口有限公司","深圳古北电子科技有限公司","南京博联智能科技有限公司"],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82412","value":"深圳古北电子科技有限公司"},{"modelname":"报销部门","isRequired":"1","isShow":"1","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82413","value":"TOB事业部"},{"modelname":"报销人","isRequired":"1","isShow":"1","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82414","value":"罗钧楠"},{"modelname":"职务","isRequired":"1","isShow":"1","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82415","value":"FAE技术支持"},{"modelname":"项目","isRequired":"0","isShow":"1","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82416","value":"公司年会，发布会，以及培训学习"},{"modelname":"预计出差时间","isRequired":"0","isShow":"1","canModifyAnytime":"0","needPrint":"0","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82417","value":"12.21"},{"modelname":"预计结束时间","isRequired":"0","isShow":"1","canModifyAnytime":"0","needPrint":"0","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82418","value":"12.27"},{"modelname":"目的地","isRequired":"1","isShow":"1","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82419","value":"杭州"},{"modelname":"出差事由","isRequired":"1","isShow":"1","canModifyAnytime":"0","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82420","value":"公司年会，发布会，以及培训学习"},{"modelname":"实际出差时间和结束时间段","isRequired":"0","isShow":"0","canModifyAnytime":"0","needPrint":"0","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82421"},{"modelname":"报销明细单","isRequired":"1","isShow":"0","canModifyAnytime":"0","modeltype":"8","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[{"columnsName":"出发或事项发生时间","columnsType":"text","$$hashKey":"object:11270"},{"columnsName":"出发地点","columnsType":"text","$$hashKey":"object:11271"},{"columnsName":"到达或事项结束时间","columnsType":"text","$$hashKey":"object:11272"},{"columnsName":"到达地点","columnsType":"text","$$hashKey":"object:11273"},{"columnsName":"交通工具","columnsType":"text","$$hashKey":"object:11274"},{"columnsName":"事项","columnsType":"text","$$hashKey":"object:11275"},{"columnsName":"交通费","columnsType":"number","$$hashKey":"object:11276"},{"columnsName":"住宿费","columnsType":"number","$$hashKey":"object:11277"},{"columnsName":"招待费","columnsType":"number","$$hashKey":"object:11278"},{"columnsName":"补贴","columnsType":"number","$$hashKey":"object:11279"},{"columnsName":"票据数","columnsType":"text","$$hashKey":"object:11280"},{"columnsName":"票据序号","columnsType":"text","$$hashKey":"object:11281"},{"columnsName":"备注","columnsType":"text","$$hashKey":"object:11282"}],"rows":[]},"needPrint":"1","$$hashKey":"object:82422","tbodyArray":[],"tableTotalArr":["","","","","","","","","","","","",""]},{"modelname":"报销总额（小写）","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82423"},{"modelname":"报销总额（大写）","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82424"},{"modelname":"超支金额","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82425"},{"modelname":"应报金额","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"1","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82426"},{"modelname":"调整金额","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"8","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[{"columnsName":"交通费","columnsType":"text","$$hashKey":"object:11422"},{"columnsName":"住宿费","columnsType":"text","$$hashKey":"object:11423"},{"columnsName":"招待费","columnsType":"text","$$hashKey":"object:11424"},{"columnsName":"补贴","columnsType":"text","$$hashKey":"object:11425"}],"rows":["调整","实报金额"]},"needPrint":"1","$$hashKey":"object:82427","tbodyArray":[[{"columnsName":"交通费","rowsName":"调整","columnsType":"text","content":""},{"columnsName":"住宿费","rowsName":"调整","columnsType":"text","content":""},{"columnsName":"招待费","rowsName":"调整","columnsType":"text","content":""},{"columnsName":"补贴","rowsName":"调整","columnsType":"text","content":""}],[{"columnsName":"交通费","rowsName":"实报金额","columnsType":"text","content":""},{"columnsName":"住宿费","rowsName":"实报金额","columnsType":"text","content":""},{"columnsName":"招待费","rowsName":"实报金额","columnsType":"text","content":""},{"columnsName":"补贴","rowsName":"实报金额","columnsType":"text","content":""}]],"tableTotalArr":["","","",""]},{"modelname":"备注","isRequired":"0","isShow":"0","canModifyAnytime":"0","modeltype":"2","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"needPrint":"1","$$hashKey":"object:82428"},{"modelname":"附件","isRequired":"0","isShow":"1","canModifyAnytime":"0","needPrint":"1","modeltype":"4","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82429","fileNameArr":[],"attachfileArr":[],"value":"[]"},{"modelname":"附单据","isRequired":"0","isShow":"0","canModifyAnytime":"0","needPrint":"1","modeltype":"4","enumList":[],"multiSelectList":[],"modeldesc":"","tableObj":{"columns":[],"rows":[]},"$$hashKey":"object:82430","fileNameArr":[],"attachfileArr":[]}]`
	json.Unmarshal([]byte(str), &templates)

	for _, template := range templates {
		isTable, key, value := handleTemplateRow(template)
		// key != "" 是过滤 Modeltype 为 "7"的情况
		if !isTable && key != "" {
			report.SetMargin(4*unit, 0)
			content := fmt.Sprintf("%s: %s", key, value)
			contentDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
			contentDiv.SetFont(textFont).SetContent(content).GenerateAtomicCellWithAutoWarp()
			report.SetMargin(0, 1*unit)

		}

		// 处理表格
		if isTable {
			report.SetMargin(4*unit, 0)
			content := fmt.Sprintf("%s:", key)
			contentDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
			contentDiv.SetFont(textFont).SetContent(content).GenerateAtomicCellWithAutoWarp()
			report.SetMargin(0, 0.5*unit)

			rows, cols, cells, hasRowName := handleTable(template)
			table := gopdf.NewTable(cols, rows, 100*unit, lineHight, report)
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					cell := table.NewCell()
					element := gopdf.NewDivWithWidth(table.GetColWithByIndex(i, j), lineHight, lineSpace, report)
					element.SetFont(textFont)
					element.SetBorder(gopdf.Scope{Left: 0.5 * unit, Top: 0.5 * unit})
					if i == 0 || j == 0 && hasRowName {
						element.SetHorizontalCentered()
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
