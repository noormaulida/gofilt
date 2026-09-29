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

// LookupOperator resolves a user-supplied operator name (word or symbol)
// to its Operator constant. The input is trimmed and lowercased before lookup.
// The second return value is false when the input is empty or unrecognized.
//
// Recognized word aliases: eq, ne, gt, gte, lt, lte, like, ilike, in, between.
// Recognized symbol aliases: =, ==, !=, <>, >, >=, <, <=.
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

// IsKnownOperator reports whether name resolves to a known operator.
// It is a convenience wrapper around LookupOperator that discards the result.
func IsKnownOperator(name string) bool {
	_, ok := LookupOperator(name)
	return ok
}
