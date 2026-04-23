Complex Markdown layout tests (blockquotes, code, lists, nesting)
================================================================

## Fenced + paragraph + nested quote (same as regression cases)

> Outer quote before code.
> 
> ``` 
> line one in outer fenced
> line two
> ```
> 
> Short paragraph after first fence.
> 
> > Inner fenced block
> > ```
> > inner.com
> > inner.io
> > ```
> 
> Back to outer: plain line after inner block.
> 
> ```
> outer-only fence
> ```
> 
> Final line in outer quote.

## Triple-nested blockquote (text only)

> L1
> > L2
> > > L3 depth
> > L2 cont
> L1 cont

## List inside blockquote

> Before list
> 
> * item A in quote
> * item B with **bold** and `code`
> 
> After list

## Indented code (4 spaces) after quote line

> Paragraph in quote.
> 
>     four-space code line A
>     four-space line B
> 
> Trailing text.

## Blockquote then heading (end section)

> End of complex tests.

---

## Mixed: hr, code, list at root

Root paragraph.

    root indented pre
    second line

* bullet
* second

`inline` and **strong** in one line.
