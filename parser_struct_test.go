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

	// Hanya izinkan 'name' dan 'status'
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
