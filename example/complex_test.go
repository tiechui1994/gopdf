package example

import (
	"fmt"
	"time"
	"testing"
	"github.com/skip2/go-qrcode"
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

func JobDetailReport() {
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

	r.RegisterExecutor(core.Executor(JobDetailReportExecutor), core.Detail)

	r.Execute("example.pdf")
	r.SaveAtomicCellText("example.txt")
}

func JobDetailReportExecutor(report *core.Report) {
	var (
		data      JobExportDetail
		unit      = report.GetUnit()
		lineSpace = 0.01 * unit
		lineHight = 1.5 * unit
	)

	ret, errStr := getExportJobData(&data)
	if ret != 0 {
		panic(struct {
			ret    int
			errStr string
		}{ret: ret, errStr: errStr})
	}

	// todo: 任务详情
	div := gopdf.NewDivWithWidth(20*unit, lineHight, lineSpace, report)
	div.SetFont(largeFont)
	div.SetContent("任务详情").GenerateAtomicCellWithAutoWarp()

	// 二维码
	//dir, _ := filepath.Abs("temp//")
	//qrcodeFile := fmt.Sprintf("%v/qrcode-%s.png", dir, report.Vars["JobId"])
	//ret, errStr = generateQrcode(qrcodeFile, "https://www.baidu.com")
	//if ret != 0 {
	//	panic(struct {
	//		ret    int
	//		errStr string
	//	}{ret: ret, errStr: errStr})
	//}
	//
	//im := gopdf.NewImage(qrcodeFile, report)
	//im.GenerateAtomicCell()

	// 基本信息
	report.SetMargin(2*unit, 1*unit)
	baseInfoDiv := gopdf.NewDivWithWidth(20*unit, lineHight, lineSpace, report)
	baseInfoDiv.SetFont(headFont)
	baseInfoDiv.SetContent("基本信息")
	baseInfoDiv.GenerateAtomicCellWithAutoWarp()
	report.SetMargin(0, 0.5*unit)

	baseInfo := gopdf.NewDivWithWidth(20*unit, lineHight, lineSpace, report)
	baseInfo.SetMarign(gopdf.Scope{Left: 4 * unit})
	baseInfo.SetFont(textFont).SetContent(fmt.Sprintf("任务名称: %s", data.JobName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("创建人: %s", data.CreatUserName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("状态: %s", data.Status)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("任务分类: %s", data.IssueClassName)).GenerateAtomicCellWithAutoWarp()
	baseInfo.CopyWithContent(fmt.Sprintf("任务: %s", data.IssueSubClassName)).GenerateAtomicCellWithAutoWarp()

	// 模板
	report.SetMargin(2*unit, 1*unit)
	templateDiv := baseInfoDiv.CopyWithContent("模板信息")
	templateDiv.GenerateAtomicCellWithAutoWarp()

	// todo: 评论
	report.SetMargin(0, 1*unit)
	commentDiv := div.CopyWithContent("评论")
	commentDiv.GenerateAtomicCellWithAutoWarp()
	report.SetMargin(0, 1*unit)

	if len(data.Contents) == 0 {
		nodataDiv := gopdf.NewDivWithWidth(50*unit, lineHight, lineSpace, report)
		nodataDiv.SetFont(textFont).SetContent("\t没有回复记录").GenerateAtomicCellWithAutoWarp()
	}
	for _, content := range data.Contents {
		cellStr := fmt.Sprintf("\t%s    %s    %s", content.Time, content.Msg, content.CreateUser)
		comment := gopdf.NewDivWithWidth(50*unit, lineHight, lineSpace, report)
		comment.SetFont(textFont).SetContent(cellStr).GenerateAtomicCellWithAutoWarp()
	}

	// todo: 历史记录
	report.SetMargin(0, 1*unit)
	historyDiv := div.CopyWithContent("历史")
	historyDiv.GenerateAtomicCellWithAutoWarp()
	report.SetMargin(0, 1*unit)

	if len(data.Contents) == 0 {
		nodataDiv := gopdf.NewDivWithWidth(50*unit, lineHight, lineSpace, report)
		nodataDiv.SetFont(textFont).SetContent("\t没有历史记录").GenerateAtomicCellWithAutoWarp()
	}

	for _, content := range data.History {
		cellStr := fmt.Sprintf("\t%s    %s    %s", content.Time, content.Msg, content.CreateUser)
		comment := gopdf.NewDivWithWidth(50*unit, lineHight, lineSpace, report)
		comment.SetFont(textFont).SetContent(cellStr).GenerateAtomicCellWithAutoWarp()
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

	AD string
	OP string
	CC []string

	Template map[string]string

	Contents []JobExportInfo
	History  []JobExportInfo
}

func getExportJobData(data *JobExportDetail) (ret int, errStr string) {
	data.JobName = "YY"
	data.CreatedAt = time.Now().Format(DateFormat)
	data.CreatUserName = "whf"
	data.Status = "ENDING"
	data.IssueClassName = "测试大类二"
	data.IssueSubClassName = "测试多行非必填文本"
	data.TimeOut = "24小时哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈"

	data.Template = map[string]string{
		"A": "哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈" +
			"哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈",
		"B": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"C": "1234567890123456789012345678901234567890" +
			"1234567890123456789012345678901234567890",
		"D": "UTF-8:Unicode Transformation Format-8bit,允许含\nBOM,但通常不含BOM.是用以解决" +
			"国际上字符的一种多字节编码,它对英文使用8位(即一个字节),中文使用24位(三个字节)来编码.UTF-8包含全世" +
			"界所有国家需要用到的字符,是国际编码,通用性强.UTF-8编码的文字可以在各国支持UTF8字符集的浏览器上显示.如,如" +
			"果是UTF8编码,则在外国人的英文IE上也能显示中文,他们无需下载IE的中文语言支持" +
			"包国家标准GB2312基础上扩容后兼容GB2312的标准.GBK的文字编码是用双字节来表示的,即不论中,英" +
			"文字符均使用双字节来表示,为了区分中文,将其最高位都设定成1.GBK包含全部中文字符,是国家编码,通用性比" +
			"UTF8差,不过UTF8占用的数据库比GBD大",
		"E": "UTF-8:Unicode Transformation Format-8bit,允许含BOM,但通常不含BOM.是用以解决" +
			"国际上字符的一种多字节编码,它对英文使用8位(即一个字节),中文使用24位(三个字节)来编码.UTF-8包含全世" +
			"界所有国家需要用到的字符,是国际编码,通用性强.UTF-8编码的文字可以在各国支持UTF8字符集的浏览器上显示.如,如" +
			"果是UTF8编码,则在外国人的英文IE上也能显示中文,他们无需下载IE的中文语言支持" +
			"包国家标准GB2312基础上扩容后兼容GB2312的标准.GBK的文字编码是用双字节来表示的,即不论中,英" +
			"文字符均使用双字节来表示,为了区分中文,将其最高位都设定成1.GBK包含全部中文字符,是国家编码,通用性比" +
			"UTF8差,不过UTF8占用的数据库比GBD大",
	}

	data.AD = "www"
	data.OP = "www"
	data.CC = []string{"www"}

	data.Contents = []JobExportInfo{
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"Hello World 包国家标准GB2312基础上扩容后兼容GB2312的标准包国家标准GB2312基础上扩容后兼容GB2312的标准包国家标准GB2312基础上扩容后兼容GB2312的标准",
			"WWW",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"How are you 包国家标准GB2312基础上扩容后兼容GB2312的标准包国家标准GB2312基础上扩容后兼容GB2312的标准包国家标准GB2312基础上扩容后兼容GB2312的标准包国家标准GB2312基础上扩容后兼容GB2312的标准",
			"WWW",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"我去我看了情况情况外壳切进去看见我看见看见请我看叫我去看进去我卡进去我看起我看我空间我卡进去我请大家看看我就去考进去即可看见我的况且我看看去我看我的情况我看到我看看去玩你去问问看请我去的胃口",
			"刘永丽",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"How are you",
			"WWW",
		},
	}

	data.History = []JobExportInfo{
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"Hello World",
			"WWW",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"How are you",
			"UTF-8:Unicode Transformation Format-8bit,允许含BOM,但通常不含BOM.是用以解决" +
				"国际上字符的一种多字节编码,它对英文使用8位(即一个字节),中文使用24位(三个字节)来编码.UTF-8包含全世" +
				"界所有国家需要用到的字符,是国际编码,通用性强.UTF-8编码的文字可以在各国支持UTF8字符集的浏览器上显示.如,如" +
				"果是UTF8编码,则在外国人的英文IE上也能显示中文,他们无需下载IE的中文语言支持" +
				"包国家标准GB2312基础上扩容后兼容GB2312的标准.GBK的文字编码是用双字节来表示的,即不论中,英" +
				"文字符均使用双字节来表示,为了区分中文,将其最高位都设定成1.GBK包含全部中文字符,是国家编码,通用性比" +
				"UTF8差,不过UTF8占用的数据库比GBD大",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"How are you",
			"WWW",
		},
		{
			time.Now().Format("2006-01-02 15:03:04"),
			"How are you",
			"UTF-8:Unicode Transformation Format-8bit,允许含BOM,但通常不含BOM.是用以解决" +
				"国际上字符的一种多字节编码,它对英文使用8位(即一个字节),中文使用24位(三个字节)来编码.UTF-8包含全世" +
				"界所有国家需要用到的字符,是国际编码,通用性强.UTF-8编码的文字可以在各国支持UTF8字符集的浏览器上显示.如,如" +
				"果是UTF8编码,则在外国人的英文IE上也能显示中文,他们无需下载IE的中文语言支持" +
				"包国家标准GB2312基础上扩容后兼容GB2312的标准.GBK的文字编码是用双字节来表示的,即不论中,英" +
				"文字符均使用双字节来表示,为了区分中文,将其最高位都设定成1.GBK包含全部中文字符,是国家编码,通用性比" +
				"UTF8差,不过UTF8占用的数据库比GBD大",
		},
	}
	return ret, errStr
}

func generateQrcode(src, content string) (ret int, errStr string) {
	err := qrcode.WriteFile(content, qrcode.Medium, 256, src)
	if err != nil {
		return ErrFile, err.Error()
	}
	return ret, errStr
}

func TestJobExport(t *testing.T) {
	JobDetailReport()
}
