package gofiltsql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/noormaulida/gofilt"
)

var safeSortField = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// BuildWHERE serializes f into a WHERE clause using PostgreSQL-style
// numbered placeholders ($1, $2, ...). Multiple conditions are joined with
// AND. It returns the clause string (prefixed with "WHERE" when non-empty)
// and the slice of arguments in the same order as the placeholders.
//
// When Filter.Conditions is empty, both return values are empty ("", nil) so
// the caller can safely concatenate without a stray WHERE keyword.
//
// LIKE and ILIKE values are wrapped with %...% automatically; all other
// operators pass the value through unchanged. OpIn and OpBetween values are
// passed as-is; callers that need placeholders expanded for these operators
// should pre-process Condition.Value or handle it separately.
func BuildWHERE(f *gofilt.Filter) (string, []any) {
	if len(f.Conditions) == 0 {
		return "", nil
	}

	var clauses []string
	var args []any

	for i, cond := range f.Conditions {
		clauses = append(clauses, fmt.Sprintf("%s %s $%d", cond.Field, cond.Operator, i+1))

		if cond.Operator == gofilt.OpLike || cond.Operator == gofilt.OpILike {
			args = append(args, fmt.Sprintf("%%%v%%", cond.Value))
		} else {
			args = append(args, cond.Value)
		}
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

// BuildORDER serializes validated sorts into an ORDER BY clause. Sorts retain
// their input order. Invalid identifiers and directions are omitted so raw
// SQL cannot be introduced through a manually constructed Filter.
//
// When Filter.Sorts is empty or contains no valid sorts, BuildORDER returns
// an empty string.
func BuildORDER(f *gofilt.Filter) string {
	orders := make([]string, 0, len(f.Sorts))
	for _, sort := range f.Sorts {
		if !safeSortField.MatchString(sort.Field) {
			continue
		}
		switch sort.Direction {
		case gofilt.DirectionAsc, gofilt.DirectionDesc:
			orders = append(orders, fmt.Sprintf("%s %s", sort.Field, sort.Direction))
		}
	}
	if len(orders) == 0 {
		return ""
	}
	return "ORDER BY " + strings.Join(orders, ", ")
}
