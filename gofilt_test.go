package gofilt

import (
	"testing"
)

func TestOperatorConstants(t *testing.T) {
	tests := []struct {
		op       Operator
		expected string
	}{
		{OpEq, "="},
		{OpNe, "!="},
		{OpGt, ">"},
		{OpGte, ">="},
		{OpLt, "<"},
		{OpLte, "<="},
		{OpLike, "LIKE"},
		{OpILike, "ILIKE"},
		{OpIn, "IN"},
		{OpBetween, "BETWEEN"},
	}

	for _, tt := range tests {
		if string(tt.op) != tt.expected {
			t.Errorf("operator constant mismatch: got %s, want %s", tt.op, tt.expected)
		}
	}
}

func TestFilterStructInitialization(t *testing.T) {
	filter := &Filter{
		Conditions: []Condition{
			{Field: "status", Operator: OpEq, Value: "active"},
		},
		Sort:   []string{"created_at DESC"},
		Limit:  10,
		Offset: 0,
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(filter.Conditions))
	}

	if filter.Conditions[0].Field != "status" {
		t.Errorf("expected field status, got %s", filter.Conditions[0].Field)
	}

	if filter.Limit != 10 {
		t.Errorf("expected limit 10, got %d", filter.Limit)
	}

	if len(filter.Sort) != 1 || filter.Sort[0] != "created_at DESC" {
		t.Errorf("unexpected sort configuration")
	}
}
