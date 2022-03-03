package ipt

type Rule struct {
	From string
	To   string
	Port string
}

type Ruleset []Rule
