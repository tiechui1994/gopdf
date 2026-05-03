package gopdf

import (
	"fmt"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/lex"
)

// markdown_blocks.go：由 lexer token 展开得到的复合结点（段落、标题、列表、引用、表格等）。

// MdFencedCode 围栏代码块：块级垂直盒模型仅在本结点首尾施加一次；子结点为 MdText + 可选 MdSpace。
type MdFencedCode struct {
	ElementBase
	blockBox        MdBoxModel
	blockTopApplied bool
	children        []markdownNode
}

func (f *MdFencedCode) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   f.pdf,
		flowColumnOffsetPt:    f.flowColumnOffsetPt,
		blockquote:            f.blockquote,
		Type:                  typ,
		hangingIndentPt:       f.hangingIndentPt,
		quoteBarsLeftOffsetPt: f.quoteBarsLeftOffsetPt,
		Margin:                f.Margin,
		Padding:               f.Padding,
		theme:                 f.theme,
	}
}

func (f *MdFencedCode) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !f.blockTopApplied {
		applyBlockTop(f.pdf, f.blockBox)
		f.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&f.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(f.pdf, f.blockBox)
	return false, true, nil
}

// MdMutiText 对应 TYPE_TEXT 带子 token 树时的扁平展开容器（名称保留历史拼写 Muti）。
type MdMutiText struct {
	ElementBase
	fonts    map[string]string
	children []markdownNode
}

// getabstract 拷贝宿主的几何与主题字段，生成子结点共用的 ElementBase（每种复合结点均有同名方法）。
func (m *MdMutiText) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   m.pdf,
		flowColumnOffsetPt:    m.flowColumnOffsetPt,
		blockquote:            m.blockquote,
		Type:                  typ,
		hangingIndentPt:       m.hangingIndentPt,
		quoteBarsLeftOffsetPt: m.quoteBarsLeftOffsetPt,
		Margin:                m.Margin,
		Padding:               m.Padding,
		theme:                 m.theme,
	}
}

func (m *MdMutiText) SetToken(t Token) error {
	if m.fonts == nil || len(m.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_TEXT {
		return fmt.Errorf("invalid type")
	}

	n := len(t.Tokens)
	for i := 0; i < n; i++ {
		token := t.Tokens[i]
		abs := m.getabstract(token.Type)
		switch token.Type {
		case TYPE_TEXT:
			if len(token.Tokens) <= 1 {
				text := &MdText{ElementBase: abs}
				txt := token.Text
				if abs.hangingIndentPt > 0 {
					txt = stripListParagraphIndent(txt)
				}
				mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
				text.SetText(m.fonts[FONT_NORMAL], txt)
				m.children = append(m.children, text)
			} else {
				mutiltext := &MdMutiText{ElementBase: abs, fonts: m.fonts}
				mutiltext.SetToken(token)
				m.children = append(m.children, mutiltext)
			}

		case TYPE_LINK:
			link := &MdText{ElementBase: abs}
			mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_LINK), &link.ElementBase)
			link.SetText(m.fonts[FONT_NORMAL], token.Text, token.Href)
			m.children = append(m.children, link)
		case TYPE_EM:
			em := &MdText{ElementBase: abs}
			mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_EM), &em.ElementBase)
			em.SetText(m.fonts[FONT_ITALIC], token.Text)
			m.children = append(m.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{ElementBase: abs}
			mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_CODESPAN), &codespan.ElementBase)
			codespan.SetText(monoFamilyFrom(m.fonts), token.Text)
			m.children = append(m.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{ElementBase: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_STRONG), &strong.ElementBase)
			strong.SetText(m.fonts[FONT_BOLD], text)
			m.children = append(m.children, strong)
		case TYPE_DEL:
			del := &MdText{ElementBase: abs}
			mergeInlineBoxModel(m.theme.BoxForInlineToken(TYPE_DEL), &del.ElementBase)
			del.SetText(m.fonts[FONT_NORMAL], token.Text)
			m.children = append(m.children, del)
		}
	}

	return nil
}

func (m *MdMutiText) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&m.children)
}

// MdHeader 标题块：上下用 MdHardBreak 施加 headingMargin*，正文为粗体字号随 depth 变化。
type MdHeader struct {
	ElementBase
	fonts           map[string]string
	children        []markdownNode
	blockBox        MdBoxModel
	blockTopApplied bool
}

// CalFontSizeAndLineHeight 按标题深度返回字号（pt）与行高（pt），并保证不低于全局最小行高。
func (h *MdHeader) CalFontSizeAndLineHeight(size int) (fontsize int, lineheight float64) {
	var fs int
	switch size {
	case 1:
		fs = 22
	case 2:
		fs = 18
	case 3:
		fs = 16
	case 4:
		fs = 13
	case 5:
		fs = 12
	case 6:
		fs = 11
	default:
		fs = 14
	}
	lh := float64(fs) * 1.38
	minLH := mdLineHeight * 1.08
	if lh < minLH {
		lh = minLH
	}
	return fs, lh
}

func (h *MdHeader) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   h.pdf,
		flowColumnOffsetPt:    h.flowColumnOffsetPt,
		blockquote:            h.blockquote,
		Type:                  typ,
		hangingIndentPt:       h.hangingIndentPt,
		quoteBarsLeftOffsetPt: h.quoteBarsLeftOffsetPt,
		Margin:                h.Margin,
		Padding:               h.Padding,
		theme:                 h.theme,
	}
}

func (h *MdHeader) SetToken(t Token) (err error) {
	if h.fonts == nil || len(h.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_HEADING {
		return fmt.Errorf("invalid type")
	}

	fontsize, lineheight := h.CalFontSizeAndLineHeight(t.Depth)
	font := core.Font{Family: h.fonts[FONT_BOLD], Size: fontsize}

	absTop := h.getabstract(TYPE_BR)
	topBrk := &MdHardBreak{ElementBase: absTop}
	topBrk.lineHeight = headingMarginTop(t.Depth)
	h.children = append(h.children, topBrk)

	// Block headings have no inline tokens; use t.Text directly.
	if len(t.Tokens) == 0 && t.Text != "" {
		abs := h.getabstract(TYPE_TEXT)
		mergeBlockHorizontalInsets(h.blockBox, &abs)
		text := &MdText{ElementBase: abs, headingStepPt: lineheight}
		mergeInlineBoxModel(h.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
		text.SetText(font, t.Text)
		h.children = append(h.children, text)
	} else {
		for _, token := range t.Tokens {
			abs := h.getabstract(token.Type)
			mergeBlockHorizontalInsets(h.blockBox, &abs)
			switch token.Type {
			case TYPE_TEXT:
				text := &MdText{ElementBase: abs, headingStepPt: lineheight}
				mergeInlineBoxModel(h.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
				text.SetText(font, token.Text)
				h.children = append(h.children, text)
			case TYPE_IMAGE:
				image := &MdImage{ElementBase: abs}
				mergeInlineBoxModel(h.theme.BoxForInlineToken(TYPE_IMAGE), &image.ElementBase)
				h.children = append(h.children, image)
			}
		}
	}

	absBot := h.getabstract(TYPE_BR)
	botBrk := &MdHardBreak{ElementBase: absBot}
	botBrk.lineHeight = headingMarginBottom(t.Depth)
	h.children = append(h.children, botBrk)

	return nil
}

func (h *MdHeader) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !h.blockTopApplied {
		applyBlockTop(h.pdf, h.blockBox)
		h.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&h.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(h.pdf, h.blockBox)
	return false, true, nil
}

// MdParagraph 段落：inline token 序列展开为 MdText / MdHardBreak / MdImage 等。
type MdParagraph struct {
	ElementBase
	fonts           map[string]string
	children        []markdownNode
	blockBox        MdBoxModel
	blockTopApplied bool
}

func (p *MdParagraph) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   p.pdf,
		flowColumnOffsetPt:    p.flowColumnOffsetPt,
		blockquote:            p.blockquote,
		Type:                  typ,
		hangingIndentPt:       p.hangingIndentPt,
		quoteBarsLeftOffsetPt: p.quoteBarsLeftOffsetPt,
		Margin:                p.Margin,
		Padding:               p.Padding,
		theme:                 p.theme,
	}
}

func (p *MdParagraph) SetToken(t Token) error {
	if p.fonts == nil || len(p.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_PARAGRAPH {
		return fmt.Errorf("invalid type")
	}

	for _, token := range t.Tokens {
		abs := p.getabstract(token.Type)
		mergeBlockHorizontalInsets(p.blockBox, &abs)
		switch token.Type {
		case TYPE_BR:
			brk := &MdHardBreak{ElementBase: abs}
			brk.lineHeight = p.theme.bodyLineHeight()
			p.children = append(p.children, brk)
		case TYPE_LINK:
			link := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_LINK), &link.ElementBase)
			link.SetText(p.fonts[FONT_NORMAL], token.Text, token.Href)
			p.children = append(p.children, link)
		case TYPE_TEXT:
			if len(token.Tokens) > 1 {
				mutiltext := &MdMutiText{ElementBase: abs, fonts: p.fonts}
				mutiltext.SetToken(token)
				p.children = append(p.children, mutiltext)
				continue
			}
			text := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
			text.SetText(p.fonts[FONT_NORMAL], token.Text)
			p.children = append(p.children, text)
		case TYPE_EM:
			em := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_EM), &em.ElementBase)
			em.SetText(p.fonts[FONT_ITALIC], token.Text)
			p.children = append(p.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_CODESPAN), &codespan.ElementBase)
			codespan.SetText(monoFamilyFrom(p.fonts), token.Text)
			p.children = append(p.children, codespan)
		case TYPE_CODE:
			code := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_CODESPAN), &code.ElementBase)
			code.SetText(monoFamilyFrom(p.fonts), token.Text)
			p.children = append(p.children, code)
		case TYPE_STRONG:
			strong := &MdText{ElementBase: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_STRONG), &strong.ElementBase)
			strong.SetText(p.fonts[FONT_BOLD], text)
			p.children = append(p.children, strong)
		case TYPE_IMAGE:
			image := &MdImage{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_IMAGE), &image.ElementBase)
			image.SetText("", token.Href)
			p.children = append(p.children, image)
		case TYPE_DEL:
			del := &MdText{ElementBase: abs}
			mergeInlineBoxModel(p.theme.BoxForInlineToken(TYPE_DEL), &del.ElementBase)
			del.SetText(p.fonts[FONT_NORMAL], token.Text)
			p.children = append(p.children, del)
		default:
			// unknown inline token; skip
		}
	}

	return nil
}

func (p *MdParagraph) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !p.blockTopApplied {
		applyBlockTop(p.pdf, p.blockBox)
		p.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&p.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(p.pdf, p.blockBox)
	return false, true, nil
}

// MdList 列表：维护 nestLevel、hangingIndentPt、markers，以及嵌套块（子列表、引用、代码）的特殊断裂结点。
type MdList struct {
	ElementBase
	fonts           map[string]string
	children        []markdownNode
	nestLevel       int
	blockBox        MdBoxModel
	blockTopApplied bool
}

func (l *MdList) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   l.pdf,
		flowColumnOffsetPt:    0, // list block offset is in hangingIndentPt only; non-zero flowColumnOffsetPt here adds fake leading spaces in GetSubText
		blockquote:            l.blockquote,
		Type:                  typ,
		hangingIndentPt:       l.hangingIndentPt,
		quoteBarsLeftOffsetPt: l.quoteBarsLeftOffsetPt,
		Margin:                l.Margin,
		Padding:               l.Padding,
		theme:                 l.theme,
	}
}

func (l *MdList) SetToken(t Token) error {
	if l.fonts == nil || len(l.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_LIST {
		return fmt.Errorf("invalid type")
	}

	l.pdf.Font(l.fonts[FONT_NORMAL], int(l.theme.bodyFontSize()), "")
	l.pdf.SetFontWithStyle(l.fonts[FONT_NORMAL], "", int(l.theme.bodyFontSize()))

	var drewMarker bool
	for index, item := range t.Items {
		drewMarker = false
		n := len(item.Tokens)

		var marker string
		if t.Ordered {
			marker = fmt.Sprintf("%d. ", index+1)
		} else {
			marker = unorderedListBulletPrefix(l.nestLevel)
		}
		// hangingIndentPt is the x-offset of the list block (e.g. inside a blockquote in a list).
		itemHang := l.hangingIndentPt + l.flowColumnOffsetPt + l.pdf.MeasureTextWidth(marker)

		for i, token := range item.Tokens {
			abs := l.getabstract(token.Type)
			mergeBlockHorizontalInsets(l.blockBox, &abs)
			abs.hangingIndentPt = itemHang

			if !drewMarker && listMarkerLeaderType(token.Type) {
				if t.Ordered {
					mabs := l.getabstract(TYPE_TEXT)
					mergeBlockHorizontalInsets(l.blockBox, &mabs)
					// Marker column = list block origin + indent padding. Use hangingIndentPt only
					// (no extra offsetx): MdSpace already places the cursor on this column for item 2+,
					// and offsetx+padding was double-counting indent after inter-item space.
					mabs.hangingIndentPt = l.hangingIndentPt + l.flowColumnOffsetPt
					mabs.flowColumnOffsetPt = 0
					text := &MdText{ElementBase: mabs, offsetx: 0, offsety: mdScale(-0.45 / 18.0)}
					mergeInlineBoxModel(l.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
					text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: int(l.theme.bodyFontSize())}, marker)
					l.children = append(l.children, text)
				}
				if !t.Ordered {
					mabs := l.getabstract(TYPE_TEXT)
					mergeBlockHorizontalInsets(l.blockBox, &mabs)
					mabs.hangingIndentPt = l.hangingIndentPt + l.flowColumnOffsetPt
					mabs.flowColumnOffsetPt = 0
					text := &MdText{ElementBase: mabs, offsetx: 0, offsety: mdScale(-0.45 / 18.0)}
					mergeInlineBoxModel(l.theme.BoxForInlineToken(TYPE_TEXT), &text.ElementBase)
					text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: int(l.theme.bodyFontSize())}, marker)
					l.children = append(l.children, text)
				}
				drewMarker = true
			}

			switch token.Type {
			case TYPE_LIST:
				nestAbs := l.getabstract(token.Type)
				mergeBlockHorizontalInsets(l.blockBox, &nestAbs)
				nestAbs.hangingIndentPt = l.hangingIndentPt
				// Align nested list with the parent item’s body column, then add one logical nest step.
				nestAbs.flowColumnOffsetPt = itemHang - l.hangingIndentPt + listNestIndentWidth()

				nest := &MdHardBreak{ElementBase: l.getabstract(TYPE_BR)}
				nest.lineHeight = l.theme.bodyLineHeight()
				// First line of nested list: X must be the nested marker column (not raw page margin),
				// since list markers no longer apply flowColumnOffsetPt via offsetx.
				nest.indentX = l.hangingIndentPt + nestAbs.flowColumnOffsetPt
				l.children = append(l.children, nest)

				sub := &MdList{ElementBase: nestAbs, fonts: l.fonts, nestLevel: l.nestLevel + 1, blockBox: l.theme.BoxList}
				sub.SetToken(token)
				l.children = append(l.children, sub)
				continue

			case TYPE_SPACE:
				gap := &MdHardBreak{ElementBase: l.getabstract(TYPE_BR)}
				gap.lineHeight = l.theme.bodyLineHeight()
				l.children = append(l.children, gap)
				continue

			case TYPE_BR:
				brk := &MdHardBreak{ElementBase: abs}
				brk.lineHeight = l.theme.bodyLineHeight()
				l.children = append(l.children, brk)
				continue

			case TYPE_BLOCKQUOTE:
				bq := &MdHardBreak{ElementBase: l.getabstract(TYPE_BR)}
				bq.lineHeight = l.theme.bodyLineHeight()
				l.children = append(l.children, bq)

				bqa := l.getabstract(TYPE_BLOCKQUOTE)
				mergeBlockHorizontalInsets(l.blockBox, &bqa)
				bqa.hangingIndentPt = itemHang
				bqa.blockquote += 1
				bqa.flowColumnOffsetPt += blockquoteIndentWidth()
				gut := mdLineHeight * (3.5 / 18.0)
				bqa.quoteBarsLeftOffsetPt = itemHang - gut
				if bqa.quoteBarsLeftOffsetPt < 0 {
					bqa.quoteBarsLeftOffsetPt = 0
				}
				blockquote := &MdBlockQuote{ElementBase: bqa, fonts: l.fonts, blockBox: l.theme.BoxBlockQuote}
				blockquote.SetToken(token)
				l.children = append(l.children, blockquote)

				if n > 0 && i == n-1 {
					for j := len(l.children) - 1; j >= 0; j-- {
						if sp, ok := l.children[j].(*MdSpace); ok {
							sp.blockquote -= 1
							break
						}
					}
				}
				continue

			case TYPE_CODE:
				cb := &MdHardBreak{ElementBase: l.getabstract(TYPE_BR)}
				cb.lineHeight = l.theme.bodyLineHeight()
				l.children = append(l.children, cb)

				cabs := l.getabstract(TYPE_CODE)
				mergeBlockHorizontalInsets(l.blockBox, &cabs)
				cabs.hangingIndentPt = itemHang
				fc := &MdFencedCode{ElementBase: cabs, blockBox: l.theme.BoxCodeBlock}
				code := &MdText{ElementBase: fc.getabstract(TYPE_CODE)}
				code.SetText(monoFamilyFrom(l.fonts), token.Text+"\n")
				fc.children = append(fc.children, code)
				ag := &MdSpace{ElementBase: fc.getabstract(TYPE_SPACE)}
				ag.lineHeight = l.theme.bodyLineHeight()
				fc.children = append(fc.children, ag)
				l.children = append(l.children, fc)
				continue
			}

			abs.hangingIndentPt = itemHang
			switch token.Type {
			case TYPE_TEXT:
				mutiltext := &MdMutiText{ElementBase: abs, fonts: l.fonts}
				mutiltext.SetToken(token)
				l.children = append(l.children, mutiltext)
				// Lexer drops single '\n' between list-item lines without a token; the next TYPE_TEXT
				// would otherwise continue on the same baseline (e.g. "...first line," + "but...").
				if i+1 < n && item.Tokens[i+1].Type == TYPE_TEXT {
					brLine := &MdHardBreak{ElementBase: l.getabstract(TYPE_BR)}
					brLine.lineHeight = l.theme.bodyLineHeight()
					l.children = append(l.children, brLine)
				}
			case TYPE_STRONG:
				strong := &MdText{ElementBase: abs}
				text := token.Text
				if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
					text = token.Tokens[0].Text
				}
				mergeInlineBoxModel(l.theme.BoxForInlineToken(TYPE_STRONG), &strong.ElementBase)
				strong.SetText(l.fonts[FONT_BOLD], text)
				l.children = append(l.children, strong)
			case TYPE_LINK:
				link := &MdText{ElementBase: abs}
				mergeInlineBoxModel(l.theme.BoxForInlineToken(TYPE_LINK), &link.ElementBase)
				link.SetText(l.fonts[FONT_NORMAL], token.Text, token.Href)
				l.children = append(l.children, link)
			}
		}

		// After an item, place the cursor on the *marker* start column for this list, not the body
		// column (itemHang). If we use itemHang here, the next item’s marker is too far right and
		// the hang-indent snap logic never runs (hangingIndentPt for marker ≠ itemHang).
		abs := l.getabstract(TYPE_SPACE)
		abs.hangingIndentPt = l.hangingIndentPt + l.flowColumnOffsetPt
		abs.blockquote -= 1
		space := &MdSpace{ElementBase: abs}
		l.children = append(l.children, space)
	}

	return nil
}

func (l *MdList) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !l.blockTopApplied {
		applyBlockTop(l.pdf, l.blockBox)
		l.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&l.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(l.pdf, l.blockBox)
	return false, true, nil
}

// MdBlockQuote 引用块：复合子结点（段落/列表/标题等）整块参与分页；blockquote 深度与 flowColumnOffsetPt 参与子结点几何。
type MdBlockQuote struct {
	ElementBase
	fonts           map[string]string
	children        []markdownNode
	blockBox        MdBoxModel
	blockTopApplied bool
}

func (b *MdBlockQuote) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   b.pdf,
		flowColumnOffsetPt:    b.flowColumnOffsetPt,
		blockquote:            b.blockquote,
		Type:                  typ,
		hangingIndentPt:       b.hangingIndentPt,
		quoteBarsLeftOffsetPt: b.quoteBarsLeftOffsetPt,
		Margin:                b.Margin,
		Padding:               b.Padding,
		theme:                 b.theme,
	}
}

func (b *MdBlockQuote) SetToken(t Token) error {
	if b.fonts == nil || len(b.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_BLOCKQUOTE {
		return fmt.Errorf("invalid type")
	}

	n := len(t.Tokens)
	for i := 0; i < n; i++ {
		token := t.Tokens[i]
		abs := b.getabstract(token.Type)
		mergeBlockHorizontalInsets(b.blockBox, &abs)
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{ElementBase: abs, fonts: b.fonts, blockBox: b.theme.BoxParagraph}
			paragraph.SetToken(token)
			b.children = append(b.children, paragraph)

			// last
			if i == n-1 {
				absSp := b.getabstract(TYPE_SPACE)
				mergeBlockHorizontalInsets(b.blockBox, &absSp)
				space := &MdSpace{ElementBase: absSp}
				space.blockquote -= 1
				b.children = append(b.children, space)
			}

		case TYPE_LIST:
			la := abs
			// Keep the same body column as blockquote text so nested list markers and lines align.
			la.hangingIndentPt = b.hangingIndentPt
			list := &MdList{ElementBase: la, fonts: b.fonts, nestLevel: 0, blockBox: b.theme.BoxList}
			list.SetToken(token)
			b.children = append(b.children, list)
		case TYPE_HEADING:
			header := &MdHeader{ElementBase: abs, fonts: b.fonts, blockBox: b.theme.BoxHeading}
			header.SetToken(token)
			b.children = append(b.children, header)
		case TYPE_BLOCKQUOTE:
			abs.blockquote += 1
			abs.flowColumnOffsetPt += blockquoteIndentWidth()
			blockquote := &MdBlockQuote{ElementBase: abs, fonts: b.fonts, blockBox: b.theme.BoxBlockQuote}
			blockquote.SetToken(token)
			b.children = append(b.children, blockquote)
		case TYPE_TEXT:
			mutiltext := &MdMutiText{ElementBase: abs, fonts: b.fonts}
			mutiltext.SetToken(token)
			b.children = append(b.children, mutiltext)
		case TYPE_SPACE:
			if i == len(t.Tokens)-1 {
				abs.blockquote -= 1
			}
			space := &MdSpace{ElementBase: abs}
			b.children = append(b.children, space)
		case TYPE_LINK:
			link := &MdText{ElementBase: abs}
			mergeInlineBoxModel(b.theme.BoxForInlineToken(TYPE_LINK), &link.ElementBase)
			link.SetText(b.fonts[FONT_NORMAL], token.Text, token.Href)
			b.children = append(b.children, link)
		case TYPE_CODE:
			fc := &MdFencedCode{ElementBase: abs, blockBox: b.theme.BoxCodeBlock}
			code := &MdText{ElementBase: fc.getabstract(TYPE_CODE)}
			code.SetText(monoFamilyFrom(b.fonts), token.Text+"\n")
			fc.children = append(fc.children, code)
			// One trailing space: gap after code (fenced with extra blank uses TYPE_TEXT as before).
			if hasBreakLine(token) {
				spa := fc.getabstract(TYPE_TEXT)
				br := &MdSpace{ElementBase: spa}
				br.lineHeight = b.theme.bodyLineHeight()
				fc.children = append(fc.children, br)
			} else {
				gap := &MdSpace{ElementBase: fc.getabstract(TYPE_SPACE)}
				gap.lineHeight = b.theme.bodyLineHeight()
				fc.children = append(fc.children, gap)
			}
			b.children = append(b.children, fc)
		case TYPE_EM:
			em := &MdText{ElementBase: abs}
			mergeInlineBoxModel(b.theme.BoxForInlineToken(TYPE_EM), &em.ElementBase)
			em.SetText(b.fonts[FONT_ITALIC], token.Text)
			b.children = append(b.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{ElementBase: abs}
			mergeInlineBoxModel(b.theme.BoxForInlineToken(TYPE_CODESPAN), &codespan.ElementBase)
			codespan.SetText(monoFamilyFrom(b.fonts), token.Text)
			b.children = append(b.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{ElementBase: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			mergeInlineBoxModel(b.theme.BoxForInlineToken(TYPE_STRONG), &strong.ElementBase)
			strong.SetText(b.fonts[FONT_BOLD], text)
			b.children = append(b.children, strong)
		case TYPE_DEL:
			del := &MdText{ElementBase: abs}
			mergeInlineBoxModel(b.theme.BoxForInlineToken(TYPE_DEL), &del.ElementBase)
			del.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, del)
		}
	}

	l := len(b.children)
	if l > 0 {
		lastType := b.children[l-1].GetType()
		if lastType != TYPE_SPACE {
			abs := b.getabstract(TYPE_TEXT)
			mergeBlockHorizontalInsets(b.blockBox, &abs)
			abs.blockquote -= 1
			br := &MdSpace{ElementBase: abs}
			b.children = append(b.children, br)
		}
	}

	return nil
}

func (b *MdBlockQuote) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !b.blockTopApplied {
		applyBlockTop(b.pdf, b.blockBox)
		b.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&b.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(b.pdf, b.blockBox)
	return false, true, nil
}

// MdTable 表格：构建 core Table + TextCell；生成 mdTableRenderer 挂入子结点队列。
type MdTable struct {
	ElementBase
	fonts           map[string]string
	children        []markdownNode
	blockBox        MdBoxModel
	blockTopApplied bool
}

func (tb *MdTable) getabstract(typ string) ElementBase {
	return ElementBase{
		pdf:                   tb.pdf,
		flowColumnOffsetPt:    tb.flowColumnOffsetPt,
		blockquote:            tb.blockquote,
		Type:                  typ,
		hangingIndentPt:       tb.hangingIndentPt,
		quoteBarsLeftOffsetPt: tb.quoteBarsLeftOffsetPt,
		Margin:                tb.Margin,
		Padding:               tb.Padding,
		theme:                 tb.theme,
	}
}

func (tb *MdTable) SetToken(t Token) error {
	if tb.fonts == nil || len(tb.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_TABLE {
		return fmt.Errorf("invalid type")
	}

	cols := len(t.Header)
	rows := len(t.Cells) + 1 // header + data rows

	if cols == 0 || rows == 0 {
		return nil
	}

	header := lex.PadAlignTo(t.Header, cols)
	align := lex.PadAlignTo(t.Align, cols)
	cells := make([][]string, len(t.Cells))
	for i := range t.Cells {
		cells[i] = lex.PadAlignTo(t.Cells[i], cols)
	}

	pageEndX, _ := tb.pdf.GetPageEndXY()
	pageStartX, _ := tb.pdf.GetPageStartXY()
	tableWidth := pageEndX - pageStartX

	table := NewTable(cols, rows, tableWidth, tb.theme.bodyLineHeight(), tb.pdf)
	table.SetMargin(core.Scope{})

	border := core.NewScope(4.0, 4.0, 4.0, 3.0)
	f := core.Font{Family: tb.fonts[FONT_BOLD], Size: int(tb.theme.bodyFontSize()), Style: ""}

	// header row
	for j, h := range header {
		cell := table.NewCell()
		tc := NewTextCell(table.GetColWidth(0, j), tb.theme.bodyLineHeight(), 1.0, tb.pdf)
		tc.SetFont(f).SetBorder(border).SetContent(strings.TrimSpace(h))
		tc.VerticalCentered()
		// apply alignment
		if j < len(align) {
			switch align[j] {
			case "center":
				tc.HorizontalCentered()
			case "right":
				tc.RightAlign()
			}
		}
		cell.SetElement(tc)
	}

	// data rows
	f = core.Font{Family: tb.fonts[FONT_NORMAL], Size: int(tb.theme.bodyFontSize()), Style: ""}
	for i, row := range cells {
		for j, val := range row {
			if j >= cols {
				break
			}
			cell := table.NewCell()
			tc := NewTextCell(table.GetColWidth(i+1, j), tb.theme.bodyLineHeight(), 1.0, tb.pdf)
			tc.SetFont(f).SetBorder(border).SetContent(strings.TrimSpace(val))
			tc.VerticalCentered()
			// apply alignment
			if j < len(align) {
				switch align[j] {
				case "center":
					tc.HorizontalCentered()
				case "right":
					tc.RightAlign()
				}
			}
			cell.SetElement(tc)
		}
	}

	tb.children = append(tb.children, &mdTableRenderer{table: table})

	return nil
}

func (tb *MdTable) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if !tb.blockTopApplied {
		applyBlockTop(tb.pdf, tb.blockBox)
		tb.blockTopApplied = true
	}
	pagebreak, over, err = CommonGenerateAtomicCell(&tb.children)
	if err != nil || pagebreak {
		return pagebreak, over, err
	}
	applyBlockBottom(tb.pdf, tb.blockBox)
	return false, true, nil
}

// mdTableRenderer 使 Table 满足 markdownNode（GenerateAtomicCell 委托 Table）。
type mdTableRenderer struct {
	table *Table
}

func (m *mdTableRenderer) SetText(font interface{}, text ...string) {}

func (m *mdTableRenderer) GetType() string {
	return TYPE_TABLE
}

func (m *mdTableRenderer) GenerateAtomicCell() (pagebreak, over bool, err error) {
	err = m.table.GenerateAtomicCell()
	return false, true, err
}

// MdHr 水平分割线（---）；前后留竖直空隙。
type MdHr struct {
	ElementBase
}

// GenerateAtomicCell 以当前基线为参考下移绘制线段（SetTokens 已跳过 HR 前的孤立 SPACE）。
func (hr *MdHr) GenerateAtomicCell() (pagebreak, over bool, err error) {
	hr.resetLayoutExtent()
	pageStartX, _ := hr.pdf.GetPageStartXY()
	pageEndX, _ := hr.pdf.GetPageEndXY()
	_, y := hr.pdf.GetXY()

	above := hr.theme.BoxHR.Margin.Top + hr.theme.BoxHR.Padding.Top
	if above <= 0 {
		above = mdLineHeight * (15.0 / 18.0)
	}
	below := hr.theme.BoxHR.Margin.Bottom + hr.theme.BoxHR.Padding.Bottom
	if below <= 0 {
		below = mdLineHeight * (12.0 / 18.0)
	}
	ruleY := y + above
	hr.noteLayoutStart(pageStartX, y)
	hr.noteLayoutExtent(pageEndX, ruleY+below)

	hr.pdf.LineType("straight", 0.5)
	hr.pdf.LineH(pageStartX, ruleY, pageEndX)

	_, pageEndY := hr.pdf.GetPageEndXY()
	spaceY := ruleY + below
	if pageEndY-spaceY < hr.theme.bodyLineHeight() {
		return true, true, nil
	}

	hr.pdf.SetXY(pageStartX, spaceY)
	return false, true, nil
}
