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
// When the filter has no expression and no conditions, both return values are
// empty ("", nil) so the caller can safely concatenate without a stray WHERE keyword.
//
// Filter.Expr, when set, is rendered instead of Filter.Conditions. Nested groups
// are parenthesized. A root AND group is not wrapped, so a flat AND list keeps
// the same shape as Filter.Conditions.
//
// LIKE and ILIKE values are wrapped with %...% automatically; all other
// operators pass the value through unchanged. OpIn and OpBetween values are
// passed as-is; callers that need placeholders expanded for these operators
// should pre-process Condition.Value or handle it separately.
// IS NULL and IS NOT NULL emit no placeholder and add no argument.
func BuildWHERE(f *gofilt.Filter) (string, []any) {
	n := 1
	if f.Expr != nil {
		clause, args := renderSQLExpr(f.Expr, &n, true)
		if clause == "" {
			return "", nil
		}
		return "WHERE " + clause, args
	}

	if len(f.Conditions) == 0 {
		return "", nil
	}

	clause, args := renderSQLExpr(gofilt.And(conditionsAsExpr(f.Conditions)...), &n, true)
	return "WHERE " + clause, args
}

func conditionsAsExpr(conditions []gofilt.Condition) []gofilt.Expression {
	items := make([]gofilt.Expression, len(conditions))
	for i, cond := range conditions {
		items[i] = cond
	}
	return items
}

func renderSQLExpr(expr gofilt.Expression, n *int, root bool) (string, []any) {
	switch v := expr.(type) {
	case gofilt.Condition:
		return renderSQLCondition(v, n)
	case *gofilt.Condition:
		if v == nil {
			return "", nil
		}
		return renderSQLCondition(*v, n)
	case gofilt.Group:
		return renderSQLGroup(v, n, root)
	case *gofilt.Group:
		if v == nil {
			return "", nil
		}
		return renderSQLGroup(*v, n, root)
	default:
		return "", nil
	}
}

func renderSQLGroup(group gofilt.Group, n *int, root bool) (string, []any) {
	op := "AND"
	if group.Operator == gofilt.LogicalOr {
		op = "OR"
	}
	var parts []string
	var args []any
	for _, item := range group.Items {
		clause, itemArgs := renderSQLExpr(item, n, false)
		if clause == "" {
			continue
		}
		parts = append(parts, clause)
		args = append(args, itemArgs...)
	}
	if len(parts) == 0 {
		return "", nil
	}
	joined := strings.Join(parts, " "+op+" ")
	if len(parts) > 1 && !(root && op == "AND") {
		joined = "(" + joined + ")"
	}
	return joined, args
}

func renderSQLCondition(cond gofilt.Condition, n *int) (string, []any) {
	if cond.Operator == gofilt.OpIsNull || cond.Operator == gofilt.OpIsNotNull {
		return fmt.Sprintf("%s %s", cond.Field, cond.Operator), nil
	}
	clause := fmt.Sprintf("%s %s $%d", cond.Field, cond.Operator, *n)
	*n++
	if cond.Operator == gofilt.OpLike || cond.Operator == gofilt.OpILike {
		return clause, []any{fmt.Sprintf("%%%v%%", cond.Value)}
	}
	return clause, []any{cond.Value}
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
