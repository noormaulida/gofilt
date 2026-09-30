package gofilt

import (
	"testing"
)

type UserQuery struct {
	Name   string `filt:"col:name;op:ilike"`
	Age    int    `filt:"col:age;op:gte"`
	Status string `filt:"col:status"` // Default operator =
	Role   string `filt:"-"`          // Ignored field
	Custom string `custom:"col:custom_field;op:ne"`
}

func TestFromStruct_Success(t *testing.T) {
	req := UserQuery{
		Name:   "Budi",
		Age:    20,
		Status: "active",
		Role:   "admin",
	}

	filter, err := FromStruct(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 3 {
		t.Fatalf("expected 3 conditions, got %d", len(filter.Conditions))
	}

	expected := []Condition{
		{Field: "name", Operator: OpILike, Value: "Budi"},
		{Field: "age", Operator: OpGte, Value: 20},
		{Field: "status", Operator: OpEq, Value: "active"},
	}

	for i, cond := range filter.Conditions {
		if cond.Field != expected[i].Field || cond.Operator != expected[i].Operator || cond.Value != expected[i].Value {
			t.Errorf("condition[%d] = %+v, expected %+v", i, cond, expected[i])
		}
	}
}

func TestFromStruct_PointerAndZeroValues(t *testing.T) {
	req := &UserQuery{
		Name: "Siti", // Age dan Status dibiarkan zero value
	}

	filter, err := FromStruct(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition for non-zero fields, got %d", len(filter.Conditions))
	}

	if filter.Conditions[0].Field != "name" || filter.Conditions[0].Value != "Siti" {
		t.Errorf("unexpected condition: %+v", filter.Conditions[0])
	}
}

func TestFromStruct_WithCustomTagName(t *testing.T) {
	req := UserQuery{
		Custom: "test_val",
	}

	filter, err := FromStruct(req, WithTagName("custom"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition with custom tag, got %d", len(filter.Conditions))
	}

	cond := filter.Conditions[0]
	if cond.Field != "custom_field" || cond.Operator != OpNe || cond.Value != "test_val" {
		t.Errorf("unexpected condition: %+v", cond)
	}
}

func TestFromStruct_WithAllowedFields(t *testing.T) {
	req := UserQuery{
		Name:   "Budi",
		Age:    25,
		Status: "active",
	}

	filter, err := FromStruct(req, WithAllowedFields("name", "status"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 2 {
		t.Fatalf("expected 2 conditions due to whitelist, got %d", len(filter.Conditions))
	}

	for _, cond := range filter.Conditions {
		if cond.Field == "age" {
			t.Errorf("field 'age' should have been filtered out by whitelist")
		}
	}
}

type QueryBadOp struct {
	Name string `filt:"col:name;op:thisisnotrealop"`
}

func TestFromStruct_UnknownOperatorDefaultsToEq(t *testing.T) {
	req := QueryBadOp{Name: "Test"}

	filter, err := FromStruct(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(filter.Conditions))
	}

	if filter.Conditions[0].Operator != OpEq {
		t.Errorf("unknown op should fall back to OpEq, got %v", filter.Conditions[0].Operator)
	}
	if filter.Conditions[0].Field != "name" {
		t.Errorf("field still required: got %q", filter.Conditions[0].Field)
	}
}

type QueryMalformedTag struct {
	Name string `filt:"col:name;thispart_just_text_no_colon;third;"`
}

func TestFromStruct_MalformedTagPartsSkipped(t *testing.T) {
	req := QueryMalformedTag{Name: "value"}

	filter, err := FromStruct(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition (col part valid, rest are skipped), got %d", len(filter.Conditions))
	}
	if filter.Conditions[0].Field != "name" {
		t.Errorf("expected col name, got %q", filter.Conditions[0].Field)
	}
	if filter.Conditions[0].Operator != OpEq {
		t.Errorf("expected default OpEq, got %v", filter.Conditions[0].Operator)
	}
}

func TestFromStruct_WithPaginationLimits(t *testing.T) {
	req := UserQuery{Name: "Budi"}

	// Test default limit applied to FromStruct
	filter, err := FromStruct(req, WithDefaultLimit(25))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.Limit != 25 {
		t.Errorf("expected limit 25, got %d", filter.Limit)
	}

	// Test max limit clamps default limit in FromStruct
	filterClamped, err := FromStruct(req, WithDefaultLimit(200), WithMaxLimit(100))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterClamped.Limit != 100 {
		t.Errorf("expected limit 100, got %d", filterClamped.Limit)
	}
}
