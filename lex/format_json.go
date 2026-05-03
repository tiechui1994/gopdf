package lex

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

// FormatTokensJSON 将语法解析树（块级 Token 切片，递归含 tokens/items/elements）格式化为 UTF-8 JSON。
// indent 非空时（例如 "  "）为可读缩进；空字符串则单行紧凑输出。
func FormatTokensJSON(tokens []Token, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	enc.SetEscapeHTML(false)
	if err := enc.Encode(tokens); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}), nil
}

// WriteTokensJSON 将语法树写入 w（无额外换行）。
func WriteTokensJSON(w io.Writer, tokens []Token, indent string) error {
	b, err := FormatTokensJSON(tokens, indent)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// PrintTokensJSON 将语法树打印到标准输出。
func PrintTokensJSON(tokens []Token, indent string) error {
	b, err := FormatTokensJSON(tokens, indent)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(b); err != nil {
		return err
	}
	if _, err := os.Stdout.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

// FormatTreeJSON 将最近一次 Lex() 保存在接收者内的语法树序列化为 JSON（等价于 FormatTokensJSON(l.tokens, indent)）。
func (l *Lexer) FormatTreeJSON(indent string) ([]byte, error) {
	return FormatTokensJSON(l.tokens, indent)
}

// LexTreeJSON 对 text 做一次 Lex，并返回其 JSON（不修改接收者的对外可见状态约定与 Lex 相同）。
func (l *Lexer) LexTreeJSON(text string, indent string) ([]byte, error) {
	tokens := l.Lex(text)
	return FormatTokensJSON(tokens, indent)
}
