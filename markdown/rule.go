package markdown

type Rule struct {
	block struct {
	}
	inline struct {
	}
}

func GetDefaultRules() *Rule {
	return &Rule{}
}
