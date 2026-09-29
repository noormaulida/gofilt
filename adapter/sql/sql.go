package gofiltsql

import (
	"fmt"
	"strings"

	"github.com/noormaulida/gofilt"
)

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
