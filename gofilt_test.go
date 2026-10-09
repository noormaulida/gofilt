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
		{OpIsNull, "IS NULL"},
		{OpIsNotNull, "IS NOT NULL"},
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
		Sorts: []Sort{
			{Field: "created_at", Direction: DirectionDesc},
		},
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

	if len(filter.Sorts) != 1 ||
		filter.Sorts[0].Field != "created_at" ||
		filter.Sorts[0].Direction != DirectionDesc {
		t.Errorf("unexpected sort configuration")
	}
}

func TestDirectionConstants(t *testing.T) {
	if DirectionAsc != "ASC" {
		t.Errorf("expected ASC, got %q", DirectionAsc)
	}
	if DirectionDesc != "DESC" {
		t.Errorf("expected DESC, got %q", DirectionDesc)
	}
}
