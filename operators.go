package main

import "regexp"

type Operator interface {
	Compare(base, target string) bool
	Symbol() string
}

var CLEANER, _ = regexp.Compile("(>=|<=|>|<|=|~|\\^)")

var Operators = []Operator{
	&Gte{
		operator: &operator{">="},
	},
	&Lte{
		operator: &operator{"<="},
	},
	&Gt{
		operator: &operator{">"},
	},
	&Lt{
		operator: &operator{"<"},
	},
	&Equal{
		operator: &operator{"="},
	},
	&Tilde{
		operator: &operator{"~"},
	},
	&Caret{
		operator: &operator{"^"},
	},
}

type operator struct {
	Operator string
}

func (opt *operator) Symbol() string {
	return opt.Operator
}

type Gte struct {
	*operator
}

func (symbol *Gte) Compare(base, target string) bool {
	return base >= target
}

type Lte struct {
	*operator
}

func (symbol *Lte) Compare(base, target string) bool {
	return base <= target
}

type Gt struct {
	*operator
}

func (symbol *Gt) Compare(base, target string) bool {
	return base > target
}

type Lt struct {
	*operator
}

func (symbol *Lt) Compare(base, target string) bool {
	return base < target
}

type Equal struct {
	*operator
}

func (symbol *Equal) Compare(base, target string) bool {
	return base == target
}

type Tilde struct {
	*operator
}

func (symbol *Tilde) Compare(base, target string) bool {
	return base >= target
}

type Caret struct {
	*operator
}

func (symbol *Caret) Compare(base, target string) bool {
	return base >= target
}

type Vers []string

func (v Vers) Len() int {
	return len(v)
}
func (v Vers) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
func (v Vers) Less(i, j int) bool {
	return CLEANER.ReplaceAllString(v[i], "") > CLEANER.ReplaceAllString(v[j], "")
}
