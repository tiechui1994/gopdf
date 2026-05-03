package gopdf

import (
	"io"

	"github.com/tiechui1994/gopdf/lex"
)

// FormatParseTreeJSON 将 lexer 输出的 Markdown 语法解析树（[]Token，含嵌套的 tokens / items / elements）格式化为 UTF-8 JSON。
// indent 非空时使用缩进（如 "  "）；空字符串为紧凑单行。
func FormatParseTreeJSON(tokens []Token, indent string) ([]byte, error) {
	return lex.FormatTokensJSON(tokens, indent)
}

// WriteParseTreeJSON 将语法解析树 JSON 写入 w。
func WriteParseTreeJSON(w io.Writer, tokens []Token, indent string) error {
	return lex.WriteTokensJSON(w, tokens, indent)
}

// PrintParseTreeJSON 将语法解析树 JSON 打印到标准输出。
func PrintParseTreeJSON(tokens []Token, indent string) error {
	return lex.PrintTokensJSON(tokens, indent)
}
