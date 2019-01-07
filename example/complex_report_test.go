package example

import (
	"fmt"
	"time"
	"testing"
	"path/filepath"

	"github.com/tiechui1994/gopdf"
	"github.com/tiechui1994/gopdf/core"
)

const (
	ErrFile    = 1
	FONT_MY    = "微软雅黑"
	FONT_MD    = "MPBOLD"
	DateFormat = "2006-01-02 15:04:05"
)

var (
	largeFont = gopdf.Font{Family: FONT_MY, Size: 15}
	headFont  = gopdf.Font{Family: FONT_MY, Size: 12}
	textFont  = gopdf.Font{Family: FONT_MY, Size: 10}
)

func ComplexReport() {
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

	r.RegisterExecutor(core.Executor(ComplexReportExecutor), core.Detail)

	r.Execute(fmt.Sprintf("complex_report_test.pdf"))
	r.SaveAtomicCellText("complex_report_test.txt")
}

func ComplexReportExecutor(report *core.Report) {
	var (
		data      JobExportDetail
		unit      = report.GetUnit()
		lineSpace = 0.01 * unit
		lineHight = 1.9 * unit
	)

	ret, errStr := getExportJobData(&data)
	if ret != 0 {
		panic(struct {
			ret    int
			errStr string
		}{ret: ret, errStr: errStr})
	}

	dir, _ := filepath.Abs("pictures")
	qrcodeFile := fmt.Sprintf("%v/qrcode.png", dir)
	line := gopdf.NewHLine(report).SetMargin(gopdf.Scope{Top: 1 * unit, Bottom: 1 * unit}).SetWidth(0.09)
	// todo: 任务详情
	div := gopdf.NewDivWithWidth(20*unit, lineHight, lineSpace, report)
	div.SetFont(largeFont)
	div.SetContent("任务详情").GenerateAtomicCellWithAutoWarp()
	line.GenerateAtomicCell()

	// 二维码
	im := gopdf.NewImageWithWidthAndHeight(qrcodeFile, 10*unit, 10*unit, report)
	im.SetMargin(gopdf.Scope{Left: 40 * unit, Top: -6 * unit})
	im.GenerateAtomicCell()

	// 基本信息
	report.SetMargin(2*unit, -4.4*unit)
	baseInfoDiv := gopdf.NewDivWithWidth(20*unit, lineHight, lineSpace, report)
	baseInfoDiv.SetFont(headFont)
	baseInfoDiv.SetContent("基本信息").GenerateAtomicCellWithAutoWarp()

	baseInfo := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
	baseInfo.SetMarign(gopdf.Scope{Left: 4 * unit, Top: 1 * unit})
	baseInfo.SetFont(textFont).SetContent(fmt.Sprintf("任务名称: %s", data.JobName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("创建人: %s", data.CreatUserName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("状态: %s", data.Status)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("任务分类: %s", data.IssueClassName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("任务: %s", data.IssueSubClassName)).GenerateAtomicCellWithAutoWarp()

	// 模板
	report.SetMargin(2*unit, 1*unit)
	baseInfoDiv.CopyWithContent("模板信息").GenerateAtomicCellWithAutoWarp()
	report.SetMargin(0, 1*unit)
	SimpleTableReportExecutor(report)

	// todo: 评论
	report.SetMargin(0, 1*unit)
	div.CopyWithContent("评论").GenerateAtomicCellWithAutoWarp()
	line.GenerateAtomicCell()

	if len(data.Contents) == 0 {
		nodataDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
		nodataDiv.SetFont(textFont).SetContent("\t没有回复记录").GenerateAtomicCellWithAutoWarp()
		report.SetMargin(0, 1*unit)
	}
	for _, content := range data.Contents {
		cellStr := fmt.Sprintf("\t%s    %s    %s", content.Time, content.Msg, content.CreateUser)
		comment := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
		comment.SetFont(textFont).SetContent(cellStr).GenerateAtomicCellWithAutoWarp()
		report.SetMargin(0, 1*unit)
	}

	// todo: 历史记录
	report.SetMargin(0, 1*unit)
	historyDiv := div.CopyWithContent("历史")
	historyDiv.GenerateAtomicCellWithAutoWarp()
	line.GenerateAtomicCell()

	if len(data.History) == 0 {
		nodataDiv := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
		nodataDiv.SetFont(textFont).SetContent("\t没有历史记录").GenerateAtomicCellWithAutoWarp()
		report.SetMargin(0, 1*unit)
	}

	for _, content := range data.History {
		cellStr := fmt.Sprintf("\t%s    %s    %s", content.Time, content.Msg, content.CreateUser)
		comment := gopdf.NewDivWithWidth(80*unit, lineHight, lineSpace, report)
		comment.SetFont(textFont).SetContent(cellStr).GenerateAtomicCellWithAutoWarp()
		report.SetMargin(0, 1*unit)
	}
}

type JobExportInfo struct {
	Time       string
	Msg        string
	CreateUser string
}

type JobExportDetail struct {
	JobName           string
	CreatedAt         string
	CreatUserName     string
	Status            string
	IssueClassName    string
	IssueSubClassName string
	TimeOut           string

	Template map[string]string

	Contents []JobExportInfo
	History  []JobExportInfo
}

func getExportJobData(data *JobExportDetail) (ret int, errStr string) {
	data.JobName = "技术指导"
	data.CreatedAt = time.Now().Format(DateFormat)
	data.CreatUserName = "钱伟长"
	data.Status = "已经完成"
	data.IssueClassName = "发动机类别"
	data.IssueSubClassName = "测试飞机发动机"

	data.Contents = []JobExportInfo{
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"涡扇发动机结构，涡扇发动机通俗的讲可以看做两根粗细不同的管子套在一起组成的。细管子里包含了低高压气机、燃烧室、低高压涡轮，最后连接至尾喷管，这根内管所包裹的空间叫做发动机的内涵道，流经里头的空气叫内涵气流。而套在内涵道外面的粗管子则包裹着风扇以及整个或者部分细管子（内涵道），我们需要注意。",
			"钱学森",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"涡扇发动机工作原理，发动机前方的风扇旋转吸入空气被分为两个部分，一部分进入细管子成为内涵气流，一部分进入粗管子成为外涵气流，外涵气流直接从发动机尾部流出形成一部分动力，而内涵气流经过压气机被压缩，成为高温高压气体，并进一步进入燃烧室被和燃油一起进一步加热膨胀冲击后面的涡轮，涡轮就像我们小时候玩的纸风车一样，被高温高压燃气带动旋转，燃气最后从尾部高速喷出，形成发动机最主要的动力。",
			"冯卡门",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"由于涡轮与风扇、压气机同在一根轴承，因此涡轮又带动了风扇、压气机一起转动。吸入空气—增压-加热-喷射—带动发动机运转，这样一个稳定的循环运动就初步建立起来，是不是看起来很像一个永动机，只要发动机一发动，就能持续的提供动力。",
			"爱因斯坦",
		},
	}

	data.History = []JobExportInfo{
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"涡扇发动机启动，其实现代的飞机尾部一般会加装了一个辅助动力装置，叫做APU.它是涡扇发动机能够启动的关键所在，APU是由一个小电动机带动的小型" +
				"燃气轮机，在APU启动后，就能源源不断的吸入空气并送到发动机后面的燃烧室燃烧然后带动整个涡扇发动机的运转，所以涡扇发动机启动前，必须先启动APU，如果" +
				"APU故障了就只能依靠地面电源车和高压气源车来实现发动机的启动",
			"邓稼先",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"很好, 照办",
			"毛泽东",
		},
	}
	return ret, errStr
}

func TestJobExport(t *testing.T) {
	ComplexReport()
}
