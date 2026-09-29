package gofiltgorm

import (
	"fmt"

	"github.com/noormaulida/gofilt"
	"gorm.io/gorm"
)

func Apply(db *gorm.DB, f *gofilt.Filter) *gorm.DB {
	for _, cond := range f.Conditions {
		switch cond.Operator {
		case gofilt.OpLike, gofilt.OpILike:
			query := fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			db = db.Where(query, fmt.Sprintf("%%%v%%", cond.Value))
		case gofilt.OpIn:
			db = db.Where(fmt.Sprintf("%s IN (?)", cond.Field), cond.Value)
		case gofilt.OpBetween:
			db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), cond.Value)
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
