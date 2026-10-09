package gofilt

import (
	"testing"
)

func TestLookupOperator_WordAliases(t *testing.T) {
	tests := []struct {
		input    string
		expected Operator
		ok       bool
	}{
		{"eq", OpEq, true},
		{"EQ", OpEq, true},
		{"Eq", OpEq, true},
		{"ne", OpNe, true},
		{"NE", OpNe, true},
		{"gt", OpGt, true},
		{"gte", OpGte, true},
		{"lt", OpLt, true},
		{"lte", OpLte, true},
		{"like", OpLike, true},
		{"LIKE", OpLike, true},
		{"ilike", OpILike, true},
		{"in", OpIn, true},
		{"between", OpBetween, true},
		{"BETWEEN", OpBetween, true},
		{"null", OpIsNull, true},
		{"isnull", OpIsNull, true},
		{"is_null", OpIsNull, true},
		{"is", OpIsNull, true},
		{"notnull", OpIsNotNull, true},
		{"isnotnull", OpIsNotNull, true},
		{"not_null", OpIsNotNull, true},
		{"is_not_null", OpIsNotNull, true},
		{"IS NOT NULL", "", false},
	}

	for _, tt := range tests {
		op, ok := LookupOperator(tt.input)
		if ok != tt.ok {
			t.Errorf("LookupOperator(%q) ok=%t, want ok=%t", tt.input, ok, tt.ok)
			continue
		}
		if ok && op != tt.expected {
			t.Errorf("LookupOperator(%q) op=%q, want %q", tt.input, op, tt.expected)
		}
	}
}

func TestLookupOperator_SymbolAliases(t *testing.T) {
	tests := []struct {
		input    string
		expected Operator
		ok       bool
	}{
		{"=", OpEq, true},
		{"==", OpEq, true},
		{"!=", OpNe, true},
		{"<>", OpNe, true},
		{">", OpGt, true},
		{">=", OpGte, true},
		{"<", OpLt, true},
		{"<=", OpLte, true},
	}

	for _, tt := range tests {
		op, ok := LookupOperator(tt.input)
		if ok != tt.ok {
			t.Errorf("LookupOperator(%q) ok=%t, want ok=%t", tt.input, ok, tt.ok)
			continue
		}
		if ok && op != tt.expected {
			t.Errorf("LookupOperator(%q) op=%q, want %q", tt.input, op, tt.expected)
		}
	}
}

func TestLookupOperator_TrimmingAndCase(t *testing.T) {
	tests := []struct {
		input    string
		expected Operator
		ok       bool
	}{
		{"  eq  ", OpEq, true},
		{"\tGTE\n", OpGte, true},
		{"  >=  ", OpGte, true},
	}

	for _, tt := range tests {
		op, ok := LookupOperator(tt.input)
		if ok != tt.ok {
			t.Errorf("LookupOperator(%q) ok=%t, want ok=%t", tt.input, ok, tt.ok)
			continue
		}
		if ok && op != tt.expected {
			t.Errorf("LookupOperator(%q) op=%q, want %q", tt.input, op, tt.expected)
		}
	}
}

func TestLookupOperator_Invalid(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\t\n",
		"unknown",
		"eqx",
		"====",
		"!!=",
	}

	for _, input := range tests {
		op, ok := LookupOperator(input)
		if ok {
			t.Errorf("LookupOperator(%q) expected ok=false, got op=%q", input, op)
		}
		if op != "" {
			t.Errorf("LookupOperator(%q) expected empty op on fail, got %q", input, op)
		}
	}
}

func TestIsKnownOperator(t *testing.T) {
	known := []string{
		"eq", "EQ", "!=", "==", "<>", ">=", "<=", "like", "ilike", "in", "between",
		"null", "isnull", "is", "notnull", "is_not_null",
	}
	for _, k := range known {
		if !IsKnownOperator(k) {
			t.Errorf("IsKnownOperator(%q) = false, want true", k)
		}
	}

	unknown := []string{"", "foobar", "eqeq", "!|="}
	for _, u := range unknown {
		if IsKnownOperator(u) {
			t.Errorf("IsKnownOperator(%q) = true, want false", u)
		}
	}
}
