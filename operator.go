package gofilt

import "strings"

var operatorAliases = map[string]Operator{
	"eq":      OpEq,
	"ne":      OpNe,
	"gt":      OpGt,
	"gte":     OpGte,
	"lt":      OpLt,
	"lte":     OpLte,
	"like":    OpLike,
	"ilike":   OpILike,
	"in":      OpIn,
	"between": OpBetween,
}

var symbolAliases = map[string]Operator{
	"=":  OpEq,
	"==": OpEq,
	"!=": OpNe,
	"<>": OpNe,
	">":  OpGt,
	">=": OpGte,
	"<":  OpLt,
	"<=": OpLte,
}

func LookupOperator(name string) (Operator, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return "", false
	}

	if op, ok := operatorAliases[key]; ok {
		return op, true
	}

	op, ok := symbolAliases[key]
	return op, ok
}

func IsKnownOperator(name string) bool {
	_, ok := LookupOperator(name)
	return ok
}
