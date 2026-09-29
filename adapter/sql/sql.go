package gofiltsql

import (
	"fmt"
	"strings"

	"github.com/noormaulida/gofilt"
)

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
