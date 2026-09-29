package gofiltgorm

import (
	"fmt"
	"reflect"

	"github.com/noormaulida/gofilt"
	"gorm.io/gorm"
)

// Apply attaches every condition in f to the GORM query db via db.Where,
// then applies Limit and Offset when they are greater than zero.
// It returns the resulting *gorm.DB so further query chaining is possible.
//
// Operator handling:
//
//	LIKE / ILIKE  -> value is wrapped with %...%
//	IN            -> value slice is expanded by GORM's IN (?) syntax
//	BETWEEN       -> value is unpacked: slice/array of len 2 becomes two args;
//	                 any other value form is passed as-is to BETWEEN ? AND ?
//	others (=, !=, >, >=, <, <=)  -> raw parameterized condition
func Apply(db *gorm.DB, f *gofilt.Filter) *gorm.DB {
	for _, cond := range f.Conditions {
		switch cond.Operator {
		case gofilt.OpLike, gofilt.OpILike:
			query := fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			db = db.Where(query, fmt.Sprintf("%%%v%%", cond.Value))
		case gofilt.OpIn:
			db = db.Where(fmt.Sprintf("%s IN (?)", cond.Field), cond.Value)
		case gofilt.OpBetween:
			lo, hi, ok := splitBetween(cond.Value)
			if ok {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), lo, hi)
			} else {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), cond.Value)
			}
		default:
			query := fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			db = db.Where(query, cond.Value)
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
