package gofilt

import (
	"net/url"
	"strconv"
	"strings"
)

var urlOpMap = map[string]Operator{
	"eq":          OpEq,
	"ne":          OpNe,
	"gt":          OpGt,
	"gte":         OpGte,
	"lt":          OpLt,
	"lte":         OpLte,
	"like":        OpLike,
	"ilike":       OpILike,
	"in":          OpIn,
	"between":     OpBetween,
	"null":        OpIsNull,
	"isnull":      OpIsNull,
	"is_null":     OpIsNull,
	"is":          OpIsNull,
	"notnull":     OpIsNotNull,
	"isnotnull":   OpIsNotNull,
	"not_null":    OpIsNotNull,
	"is_not_null": OpIsNotNull,
}

// FromURL parses an HTTP query string (url.Values) into a Filter.
//
// Field syntax is `field` (defaults to OpEq) or `field[operator]` for an
// explicit operator — e.g. name[ilike], age[gte].
//
// Special keys:
//
//	limit=<int>  -> sets Filter.Limit (validated against WithMaxLimit / WithDefaultLimit)
//	offset=<int> -> sets Filter.Offset
//	page=<int>   -> sets Filter.Offset (alias for offset)
//	sort=name,-created_at -> appends allowed ascending/descending sorts
//
// Negative limit or offset values return ErrInvalidLimit or ErrInvalidOffset.
//
// Values containing commas are automatically split into a slice and the
// operator is forced to OpIn, unless the operator was already set to OpIn.
//
// Empty values and empty query keys are ignored. Sort fields are ignored
// unless explicitly whitelisted with WithAllowedSorts.
func FromURL(values url.Values, opts ...Option) (*Filter, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	filter := &Filter{}

	hasLimit := false
	var parsedLimit int
	if valList, exists := values["limit"]; exists && len(valList) > 0 && valList[0] != "" {
		l, err := strconv.Atoi(valList[0])
		if err != nil || l < 0 {
			return nil, ErrInvalidLimit
		}
		parsedLimit = l
		hasLimit = true
	}
	filter.Limit = cfg.resolveLimit(parsedLimit, hasLimit)

	hasOffset := false
	if valList, exists := values["offset"]; exists && len(valList) > 0 && valList[0] != "" {
		o, err := strconv.Atoi(valList[0])
		if err != nil || o < 0 {
			return nil, ErrInvalidOffset
		}
		filter.Offset = o
		hasOffset = true
	}
	if valList, exists := values["page"]; exists && len(valList) > 0 && valList[0] != "" {
		p, err := strconv.Atoi(valList[0])
		if err != nil || p < 0 {
			return nil, ErrInvalidOffset
		}
		if !hasOffset {
			filter.Offset = p
		}
	}

	for key, valList := range values {
		if len(valList) == 0 || valList[0] == "" || key == "" {
			continue
		}

		if key == "limit" || key == "offset" || key == "page" {
			continue
		}

		if key == "sort" {
			for _, value := range valList {
				filter.Sorts = append(filter.Sorts, parseSorts(value, cfg)...)
			}
			continue
		}

		field, op := parseURLKey(key)
		val := valList[0]

		if !cfg.isAllowed(field) {
			continue
		}

		var parsedVal any = val
		if op != OpIsNull && op != OpIsNotNull && (op == OpIn || strings.Contains(val, ",")) {
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

func parseSorts(value string, cfg *options) []Sort {
	var sorts []Sort
	for _, raw := range strings.Split(value, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		direction := DirectionAsc
		field := raw
		if strings.HasPrefix(raw, "-") {
			direction = DirectionDesc
			field = strings.TrimSpace(strings.TrimPrefix(raw, "-"))
		}

		if field == "" || !cfg.isSortAllowed(field) {
			continue
		}
		sorts = append(sorts, Sort{Field: field, Direction: direction})
	}
	return sorts
}
