package markdown

type Options struct {
	baseUrl      string
	breaks       bool
	gfm          bool
	headerIds    bool
	headerPrefix string
	langPrefix   string
	highlight    interface{}
	mangle       bool
	pedantic     bool
	renderer     *Options
	sanitize     bool
	sanitizer    interface{}
	silent       bool
	smartLists   bool
	smartypants  bool
	tokenizer    *Tokenizer
	walkTokens   interface{}
	xhtml        bool
}

func GetDefaults() *Options {
	return &Options{
		baseUrl:      "",
		breaks:       false,
		gfm:          true,
		headerIds:    true,
		headerPrefix: "",
		highlight:    "",
		langPrefix:   "language-",
		mangle:       true,
		pedantic:     false,
		sanitize:     false,
		sanitizer:    nil,
		silent:       false,
		smartLists:   false,
		smartypants:  false,
		tokenizer:    nil,
		walkTokens:   nil,
		xhtml:        false,
	}
}
