package gofiltgorm

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/noormaulida/gofilt"
	"gorm.io/gorm"
)

var safeSortField = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// Apply attaches the filter expression in f to the GORM query db via db.Where,
// applies validated sorts in order, then applies Limit and Offset when they
// are greater than zero.
// It returns the resulting *gorm.DB so further query chaining is possible.
//
// Operator handling:
//
//	LIKE / ILIKE  -> value is wrapped with %...%
//	IN            -> value slice is expanded by GORM's IN (?) syntax
//	BETWEEN       -> value is unpacked: slice/array of len 2 becomes two args;
//	                 any other value form is passed as-is to BETWEEN ? AND ?
//	IS NULL / IS NOT NULL -> no bound value
//	others (=, !=, >, >=, <, <=)  -> raw parameterized condition
//
// Filter.Expr, when set, is applied as one expression. Nested OR and AND
// groups are parenthesized. Otherwise each Condition is applied with AND.
func Apply(db *gorm.DB, f *gofilt.Filter) *gorm.DB {
	if f.Expr != nil {
		if clause, args := renderGormExpr(f.Expr, true); clause != "" {
			db = db.Where(clause, args...)
		}
	} else {
		for _, cond := range f.Conditions {
			clause, args := renderGormCondition(cond)
			db = db.Where(clause, args...)
		}
	}

	for _, sort := range f.Sorts {
		if !safeSortField.MatchString(sort.Field) {
			continue
		}
		switch sort.Direction {
		case gofilt.DirectionAsc, gofilt.DirectionDesc:
			db = db.Order(fmt.Sprintf("%s %s", sort.Field, sort.Direction))
		}
	}

	if f.Limit > 0 {
		db = db.Limit(f.Limit)
	}
	if f.Offset > 0 {
		db = db.Offset(f.Offset)
	}

	return db
}

func renderGormExpr(expr gofilt.Expression, root bool) (string, []any) {
	switch v := expr.(type) {
	case gofilt.Condition:
		return renderGormCondition(v)
	case *gofilt.Condition:
		if v == nil {
			return "", nil
		}
		return renderGormCondition(*v)
	case gofilt.Group:
		return renderGormGroup(v, root)
	case *gofilt.Group:
		if v == nil {
			return "", nil
		}
		return renderGormGroup(*v, root)
	default:
		return "", nil
	}
}

func renderGormGroup(group gofilt.Group, root bool) (string, []any) {
	op := "AND"
	if group.Operator == gofilt.LogicalOr {
		op = "OR"
	}
	var parts []string
	var args []any
	for _, item := range group.Items {
		clause, itemArgs := renderGormExpr(item, false)
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

func renderGormCondition(cond gofilt.Condition) (string, []any) {
	switch cond.Operator {
	case gofilt.OpIsNull, gofilt.OpIsNotNull:
		return fmt.Sprintf("%s %s", cond.Field, cond.Operator), nil
	case gofilt.OpLike, gofilt.OpILike:
		return fmt.Sprintf("%s %s ?", cond.Field, cond.Operator), []any{fmt.Sprintf("%%%v%%", cond.Value)}
	case gofilt.OpIn:
		return fmt.Sprintf("%s IN (?)", cond.Field), []any{cond.Value}
	case gofilt.OpBetween:
		lo, hi, ok := splitBetween(cond.Value)
		if ok {
			return fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), []any{lo, hi}
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), []any{cond.Value}
	default:
		return fmt.Sprintf("%s %s ?", cond.Field, cond.Operator), []any{cond.Value}
	}
}

func splitBetween(v any) (lo any, hi any, ok bool) {
	if v == nil {
		return nil, nil, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Len() == 2 {
			return rv.Index(0).Interface(), rv.Index(1).Interface(), true
		}
	}
	return nil, nil, false
}
