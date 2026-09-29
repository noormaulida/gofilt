package gofilt

import (
	"net/url"
	"strconv"
	"strings"
)

var urlOpMap = map[string]Operator{
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

// FromURL parses an HTTP query string (url.Values) into a Filter.
//
// Field syntax is `field` (defaults to OpEq) or `field[operator]` for an
// explicit operator — e.g. name[ilike], age[gte].
//
// Special keys:
//
//	limit=<int>  -> sets Filter.Limit
//	offset=<int> -> sets Filter.Offset (alias "page" also works)
//
// Values containing commas are automatically split into a slice and the
// operator is forced to OpIn, unless the operator was already set to OpIn.
//
// Empty values and empty query keys are ignored.
func FromURL(values url.Values, opts ...Option) (*Filter, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	filter := &Filter{}

	for key, valList := range values {
		if len(valList) == 0 || valList[0] == "" {
			continue
		}

		if key == "limit" {
			if l, err := strconv.Atoi(valList[0]); err == nil {
				filter.Limit = l
			}
			continue
		}
		if key == "offset" || key == "page" {
			if o, err := strconv.Atoi(valList[0]); err == nil {
				filter.Offset = o
			}
			continue
		}

		field, op := parseURLKey(key)
		val := valList[0]

		if !cfg.isAllowed(field) {
			continue
		}

		var parsedVal any = val
		if op == OpIn || strings.Contains(val, ",") {
			parts := strings.Split(val, ",")
			if len(parts) > 1 {
				op = OpIn
				parsedVal = parts
			}
		}

		filter.Conditions = append(filter.Conditions, Condition{
			Field:    field,
			Operator: op,
			Value:    parsedVal,
		})
	}

	return filter, nil
}

func parseURLKey(key string) (string, Operator) {
	openIdx := strings.Index(key, "[")
	closeIdx := strings.Index(key, "]")

	if openIdx != -1 && closeIdx != -1 && closeIdx > openIdx {
		field := key[:openIdx]
		rawOp := strings.ToLower(key[openIdx+1 : closeIdx])

		if op, exists := urlOpMap[rawOp]; exists {
			return field, op
		}
		return field, OpEq
	}

	return key, OpEq
}
