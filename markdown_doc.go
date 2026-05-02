package gopdf

// markdown_doc.go：从 lexer Token 切片构建复合 markdownNode 列表并驱动分页绘制。

import (
	"fmt"

	"github.com/tiechui1994/gopdf/core"
)

// MarkdownText 将 Markdown lex 结果排版到 Report：SetTokens 保留段落/列表等复合结点，GenerateAtomicCell 顺序绘制。
type MarkdownText struct {
	quote       bool
	pdf         *core.Report
	fonts       map[string]string
	theme       MarkdownTheme
	children    []mardown
	x           float64
	writedLines int
}

// NewMarkdownText 创建渲染器：fonts 须包含 FONT_NORMAL / FONT_BOLD / FONT_IALIC；默认主题为 DefaultMarkdownTheme。
// x 为允许的左侧起始坐标下界（小于页左边距时会被抬升到页起点）。
func NewMarkdownText(pdf *core.Report, x float64, fonts map[string]string) (*MarkdownText, error) {
	px, _ := pdf.GetPageStartXY()
	if x < px {
		x = px
	}

	if fonts == nil || fonts[FONT_BOLD] == "" || fonts[FONT_IALIC] == "" || fonts[FONT_NORMAL] == "" {
		return nil, fmt.Errorf("invalid fonts")
	}

	mt := MarkdownText{
		pdf:   pdf,
		x:     x,
		fonts: fonts,
		theme: DefaultMarkdownTheme(),
	}

	return &mt, nil
}

// WithTheme 覆盖文档级 MarkdownTheme（按值拷贝）。
func (mt *MarkdownText) WithTheme(t MarkdownTheme) *MarkdownText {
	mt.theme = t
	return mt
}

// getabstract 构造顶层 ElementBase（仅 pdf/type/theme；各块级类型的 Margin/Padding 由 MarkdownTheme 中对应 Box* 在构造结点时注入）。
func (mt *MarkdownText) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:   mt.pdf,
		Type:  typ,
		theme: mt.theme,
	}
}

// SetTokens 将 lexer 输出转换为结点树：每个段落/列表/标题/引用/表格/围栏代码为单个 markdownNode。
func (mt *MarkdownText) SetTokens(tokens []Token) {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		abs := mt.getabstract(token.Type)

		if token.Type == TYPE_SPACE && i+1 < len(tokens) && tokens[i+1].Type == TYPE_HR {
			continue
		}

		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{ElementBase: abs, fonts: mt.fonts, blockBox: mt.theme.BoxParagraph}
			paragraph.SetToken(token)
			mt.children = append(mt.children, paragraph)
		case TYPE_LIST:
			la := abs
			la.listHangIndent = 0
			list := &MdList{ElementBase: la, fonts: mt.fonts, nestLevel: 0, blockBox: mt.theme.BoxList}
			list.SetToken(token)
			mt.children = append(mt.children, list)
		case TYPE_HEADING:
			header := &MdHeader{ElementBase: abs, fonts: mt.fonts, blockBox: mt.theme.BoxHeading}
			header.SetToken(token)
			mt.children = append(mt.children, header)
		case TYPE_BLOCKQUOTE:
			abs.blockquote = 1
			abs.FlowInset += blockquoteIndentWidth()
			blockquote := &MdBlockQuote{ElementBase: abs, fonts: mt.fonts, blockBox: mt.theme.BoxBlockQuote}
			blockquote.SetToken(token)
			mt.children = append(mt.children, blockquote)
		case TYPE_TEXT:
			mutiltext := &MdMutiText{ElementBase: abs, fonts: mt.fonts}
			mutiltext.SetToken(token)
			mt.children = append(mt.children, mutiltext)
		case TYPE_SPACE:
			space := &MdSpace{ElementBase: abs}
			mt.children = append(mt.children, space)
		case TYPE_LINK:
			link := &MdText{ElementBase: abs}
			mergeInlineBoxModel(mt.theme.BoxForInlineToken(TYPE_LINK), &link.ElementBase)
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.children = append(mt.children, link)
		case TYPE_CODE:
			abs.Type = TYPE_CODE
			abs.FlowInset = mdScale(15.0 / 18.0)
			mergeBlockHorizontalInsets(mt.theme.BoxCodeBlock, &abs)
			fc := &MdFencedCode{ElementBase: abs, blockBox: mt.theme.BoxCodeBlock}
			code := &MdText{ElementBase: fc.getabstract(TYPE_CODE)}
			code.SetText(monoFamilyFrom(mt.fonts), token.Text+"\n")
			fc.children = append(fc.children, code)
			spAbs := fc.getabstract(TYPE_SPACE)
			spAbs.lineHeight = codeBlockAfterGap()
			space := &MdSpace{ElementBase: spAbs}
			fc.children = append(fc.children, space)
			mt.children = append(mt.children, fc)
		case TYPE_BR:
			brk := &MdHardBreak{ElementBase: abs}
			brk.lineHeight = mt.theme.bodyLineHeight()
			mt.children = append(mt.children, brk)
		case TYPE_EM:
			em := &MdText{ElementBase: abs}
			mergeInlineBoxModel(mt.theme.BoxForInlineToken(TYPE_EM), &em.ElementBase)
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.children = append(mt.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{ElementBase: abs}
			mergeInlineBoxModel(mt.theme.BoxForInlineToken(TYPE_CODESPAN), &codespan.ElementBase)
			codespan.SetText(monoFamilyFrom(mt.fonts), token.Text)
			mt.children = append(mt.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{ElementBase: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			mergeInlineBoxModel(mt.theme.BoxForInlineToken(TYPE_STRONG), &strong.ElementBase)
			strong.SetText(mt.fonts[FONT_BOLD], text)
			mt.children = append(mt.children, strong)
		case TYPE_DEL:
			del := &MdText{ElementBase: abs}
			mergeInlineBoxModel(mt.theme.BoxForInlineToken(TYPE_DEL), &del.ElementBase)
			del.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, del)
		case TYPE_TABLE:
			tb := &MdTable{ElementBase: abs, fonts: mt.fonts, blockBox: mt.theme.BoxTable}
			tb.SetToken(token)
			mt.children = append(mt.children, tb)
		case TYPE_HR:
			hr := &MdHr{ElementBase: abs}
			mt.children = append(mt.children, hr)

			abs2 := mt.getabstract(TYPE_SPACE)
			space := &MdSpace{ElementBase: abs2}
			mt.children = append(mt.children, space)
		}
	}
}

// GenerateAtomicCell 依次绘制 children；pagebreak 时 BreakPage 且若 over 则跳过已完成结点。
func (mt *MarkdownText) GenerateAtomicCell() (err error) {
	if len(mt.children) == 0 {
		return fmt.Errorf("not set text")
	}

	lc := NewLayoutContext(mt.pdf)
	for i := 0; i < len(mt.children); {
		child := mt.children[i]

		pagebreak, over, err := child.GenerateAtomicCell()
		if err != nil {
			i++
			continue
		}

		if pagebreak {
			if over {
				i++
			}
			lc.BreakPage()
			continue
		}

		if over {
			i++
		}
	}

	return nil
}
