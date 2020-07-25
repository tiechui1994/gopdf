package gopdf

import (
	"testing"

	"github.com/tiechui1994/gopdf/core"
	"io/ioutil"
	"regexp"
	"strings"
	"log"
)

const (
	MD_IG = "IPAexG"
	MD_MC = "Microsoft"
	MD_MB = "Microsoft Bold"
)

func MdReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: MD_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: MD_MC,
		FileName: "example//ttf/microsoft.ttf",
	}
	font3 := core.FontMap{
		FontName: MD_MB,
		FileName: "example//ttf/microsoft-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(MdReportExecutor), core.Detail)

	r.Execute("markdown_test.pdf")
	r.SaveAtomicCellText("markdown_test.txt")
}

/*
> 注意:
>
> 1.在链接C库的使用, 不支持条件选择. 并且CGO参数有严格的格式 `#cgo CFLAGS:...` 或者
> `#cgo LDFLAGS: ... `, 即 `#cgo` 和 参数(`CFLAGS`, `LDFLAGS`)
>
> 2.对于C语言库(`.h` 文件定义内容 和 `.c` 文件实现 `.h` 的定义), 在CGO当中引用 `.h` 文件, 必须采用
> `动态库/静态库` 链接的方式, 否则可能无法编译通过.
> **dd**, `**dd**`, `vivo`, qkwqkwwq,jdjjsdjsd, `**sss**` **`ss`**
>
>
*/

func MdReportExecutor(report *core.Report) {
	text, _ := ioutil.ReadFile("./md.md")
	mt, _ := NewMarkdownText(report, 10, map[string]string{
		FONT_NORMAL: MD_MC,
		FONT_IALIC:  MD_MC,
		FONT_BOLD:   MD_MC,
	})
	mt.SetText(string(text))
	mt.GenerateAtomicCell()
	return
	ctetx := content{
		pdf:  report,
		Type: TYPE_NORMAL,
	}
	ctetx.SetText(MD_MC, "中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888")

	btetx := content{
		pdf:  report,
		Type: TYPE_WARP,
	}
	btetx.SetText(MD_MC, "中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888中田田日日000088888")
	itetx := content{
		pdf:  report,
		Type: TYPE_NORMAL,
	}
	ctetx.GenerateAtomicCell()
	itetx.SetText(MD_MB, "中田田日日000088888空间卡生萨斯卡拉斯科拉11209901接口就爱看手机爱卡就爱29912-0-接口就爱看手机爱卡就爱021-012-12012-021-012-0-021-0120-1201221")
	btetx.GenerateAtomicCell()
	itetx.GenerateAtomicCell()
}

func TestMd(t *testing.T) {
	MdReport()
}

func TestReSubText(t *testing.T) {
	data, _ := ioutil.ReadFile("./md.md")
	text := string(data)

	matched := rensort.FindString(text)
	t.Log("matched", matched)

	rebreak := regexp.MustCompile(`\n\s*\n`)
	if rebreak.MatchString(matched) {
		index := rebreak.FindAllStringIndex(matched, 1)
		t.Log("index", matched[:index[0][0]], matched[index[0][0]] == '\n')
	}
}

func TestA(t *testing.T) {
	InitFunc()
	for k, v := range block {
		if strings.HasPrefix(k, "_") {
			continue
		}
		log.Println("key", k)
		log.Println("regex", v)
		log.Println()
	}
}
