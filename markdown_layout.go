package gopdf

import "github.com/tiechui1994/gopdf/core"

// markdown_layout.go：复合结点子结点队列上的分页与切片策略（CommonGenerateAtomicCell）。

// applyBlockTop 在块级结点首次绘制子结点之前下移光标，仅施加盒模型的上边距与上内边距（水平Inset 在 SetToken 时并入子结点 Margin，避免跨页后丢失）。
func applyBlockTop(r *core.Report, box MdBoxModel) {
	if r == nil {
		return
	}
	x, y := r.GetXY()
	y += box.Margin.Top + box.Padding.Top
	r.SetXY(x, y)
}

// applyBlockBottom 在块级结点全部子结点绘制完毕后施加下边距与下内边距。
func applyBlockBottom(r *core.Report, box MdBoxModel) {
	if r == nil {
		return
	}
	x, y := r.GetXY()
	y += box.Margin.Bottom + box.Padding.Bottom
	r.SetXY(x, y)
}

// mergeBlockHorizontalInsets 将块级盒模型的左右 Margin/Padding 累加到 ElementBase（竖直方向由 applyBlockTop/Bottom 处理）。
func mergeBlockHorizontalInsets(box MdBoxModel, e *ElementBase) {
	if e == nil {
		return
	}
	e.Margin.Left += box.Margin.Left + box.Padding.Left
	e.Margin.Right += box.Margin.Right + box.Padding.Right
}

// mergeInlineBoxModel 将行内类型对应的盒模型并入 ElementBase（各边相加）。
func mergeInlineBoxModel(ib MdBoxModel, e *ElementBase) {
	if e == nil {
		return
	}
	e.Margin.Top += ib.Margin.Top
	e.Margin.Right += ib.Margin.Right
	e.Margin.Bottom += ib.Margin.Bottom
	e.Margin.Left += ib.Margin.Left
	e.Padding.Top += ib.Padding.Top
	e.Padding.Right += ib.Padding.Right
	e.Padding.Bottom += ib.Padding.Bottom
	e.Padding.Left += ib.Padding.Left
}

// CommonGenerateAtomicCell 顺序执行 children 中每个 markdownNode。
//
// 与分页相关的切片约定（须与顶层 MarkdownText.GenerateAtomicCell 索引策略一致）：
//   - 若第 i 个子结点返回 pagebreak==true 且 over==true，且后面仍有兄弟：剪掉 [0,i]（含当前结点），
//     下一页从原 i+1 继续（典型：MdSpace 页底放不下时被丢弃）。
//   - 其它 pagebreak：保留切片从 i 起；若 over==false，同一 MdText 等会在下一页从头结点继续。
func CommonGenerateAtomicCell(children *[]mardown) (pagebreak, over bool, err error) {
	for i, comment := range *children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(*children)-1 {
				*children = (*children)[i+1:]
				return pagebreak, len(*children) == 0, nil
			}

			*children = (*children)[i:]
			return pagebreak, len(*children) == 0, nil
		}
	}
	return false, true, nil
}
