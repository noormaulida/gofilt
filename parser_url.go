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
//	or.<field>[op]=value  -> adds the predicate to one OR group under the root AND
//	or.<n>.<field>[op]=value -> adds the predicate to indexed OR group n
//	and.<n>.<field>[op]=value -> adds the predicate to indexed AND group n
//
// Negative limit or offset values return ErrInvalidLimit or ErrInvalidOffset.
//
// Values containing commas are automatically split into a slice and the
// operator is forced to OpIn, unless the operator was already set to OpIn.
//
// Empty values and empty query keys are ignored. Sort fields are ignored
// unless explicitly whitelisted with WithAllowedSorts.
//
// URL values stay strings unless WithFieldTypes registers a type for the
// column. Registered values are parsed before the condition is appended.
// A value that does not match its registered type returns ErrInvalidValue.
//
// Keys prefixed with or. or and. build Filter.Expr. Other fields stay an
// implicit AND in Filter.Conditions. A query with no grouped keys leaves
// Expr nil.
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

	groups := map[string]*Group{}
	var groupOrder []string

	for key, valList := range values {
		if len(valList) == 0 || key == "" {
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

		logical, groupID, fieldKey, grouped := parseLogicalKey(key)
		if !grouped {
			if valList[0] == "" {
				continue
			}
			cond, err := parseURLCondition(key, valList[0], cfg)
			if err != nil {
				return nil, err
			}
			if cond == nil {
				continue
			}
			filter.Conditions = append(filter.Conditions, *cond)
			continue
		}

		groupKey := string(logical) + ":" + groupID
		group := groups[groupKey]
		if group == nil {
			group = &Group{Operator: logical}
			groups[groupKey] = group
			groupOrder = append(groupOrder, groupKey)
		}
		for _, raw := range valList {
			if raw == "" {
				continue
			}
			cond, err := parseURLCondition(fieldKey, raw, cfg)
			if err != nil {
				return nil, err
			}
			if cond == nil {
				continue
			}
			group.Items = append(group.Items, *cond)
		}
	}

	filter.Expr = buildURLExpr(filter.Conditions, groups, groupOrder)
	return filter, nil
}

func parseURLCondition(key, val string, cfg *options) (*Condition, error) {
	field, op := parseURLKey(key)
	if !cfg.isAllowed(field) {
		return nil, nil
	}

	var parsedVal any = val
	inParts := []string(nil)
	if op != OpIsNull && op != OpIsNotNull && (op == OpIn || strings.Contains(val, ",")) {
		parts := strings.Split(val, ",")
		if len(parts) > 1 {
			op = OpIn
			inParts = parts
			parsedVal = parts
		}
	}

	if typ, ok := cfg.fieldTypes[field]; ok && op != OpIsNull && op != OpIsNotNull {
		converted, err := convertURLValue(parsedVal, inParts, typ)
		if err != nil {
			return nil, err
		}
		parsedVal = converted
	}

	return &Condition{Field: field, Operator: op, Value: parsedVal}, nil
}

func parseLogicalKey(key string) (op LogicalOperator, id, fieldKey string, ok bool) {
	switch {
	case strings.HasPrefix(key, "or."):
		op = LogicalOr
		key = strings.TrimPrefix(key, "or.")
	case strings.HasPrefix(key, "and."):
		op = LogicalAnd
		key = strings.TrimPrefix(key, "and.")
	default:
		return "", "", "", false
	}
	if key == "" {
		return "", "", "", false
	}
	if i := strings.IndexByte(key, '.'); i > 0 && digitsOnly(key[:i]) {
		id = key[:i]
		key = key[i+1:]
		if key == "" {
			return "", "", "", false
		}
	}
	return op, id, key, true
}

func digitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func buildURLExpr(conditions []Condition, groups map[string]*Group, order []string) Expression {
	var grouped []Expression
	for _, key := range order {
		group := groups[key]
		if group == nil || len(group.Items) == 0 {
			continue
		}
		grouped = append(grouped, *group)
	}
	if len(grouped) == 0 {
		return nil
	}
	if len(conditions) == 0 && len(grouped) == 1 {
		return grouped[0]
	}
	items := make([]Expression, 0, len(conditions)+len(grouped))
	for _, cond := range conditions {
		items = append(items, cond)
	}
	items = append(items, grouped...)
	return And(items...)
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
