package gofilt

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
)

func TestAndOrHelpers(t *testing.T) {
	tree := And(
		Condition{Field: "status", Operator: OpEq, Value: "active"},
		Condition{Field: "age", Operator: OpGte, Value: 18},
		Or(
			Condition{Field: "role", Operator: OpEq, Value: "admin"},
			Condition{Field: "role", Operator: OpEq, Value: "moderator"},
		),
	)
	if tree.Operator != LogicalAnd || len(tree.Items) != 3 {
		t.Fatalf("and group: %+v", tree)
	}
	or, ok := tree.Items[2].(Group)
	if !ok || or.Operator != LogicalOr || len(or.Items) != 2 {
		t.Fatalf("or group: %#v", tree.Items[2])
	}
	var expr Expression = tree
	if expr == nil {
		t.Fatal("group should be an expression")
	}
	Condition{}.expression()
	Group{}.expression()
}

func TestFromURL_OrGroup(t *testing.T) {
	query := url.Values{
		"status":   []string{"active"},
		"age[gte]": []string{"18"},
		"or.role":  []string{"admin", "moderator"},
	}
	filter, err := FromURL(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(filter.Conditions) != 2 {
		t.Fatalf("flat conditions: %d", len(filter.Conditions))
	}
	root, ok := filter.Expr.(Group)
	if !ok || root.Operator != LogicalAnd || len(root.Items) != 3 {
		t.Fatalf("expr: %#v", filter.Expr)
	}
	or, ok := root.Items[2].(Group)
	if !ok || or.Operator != LogicalOr || len(or.Items) != 2 {
		t.Fatalf("or items: %#v", root.Items[2])
	}
	if or.Items[0].(Condition).Value != "admin" || or.Items[1].(Condition).Value != "moderator" {
		t.Fatalf("or values: %#v", or.Items)
	}
}

func TestFromURL_IndexedGroups(t *testing.T) {
	query := url.Values{
		"or.0.role":       []string{"admin", ""},
		"or.1.status[ne]": []string{"guest"},
		"and.0.age[gte]":  []string{"18"},
	}
	filter, err := FromURL(query, WithAllowedFields("role", "status", "age"))
	if err != nil {
		t.Fatal(err)
	}
	root, ok := filter.Expr.(Group)
	if !ok || root.Operator != LogicalAnd {
		t.Fatalf("expr: %#v", filter.Expr)
	}
	if len(root.Items) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(root.Items))
	}
}

func TestFromURL_SingleOrGroup(t *testing.T) {
	filter, err := FromURL(url.Values{"or.role": []string{"admin", "moderator"}})
	if err != nil {
		t.Fatal(err)
	}
	or, ok := filter.Expr.(Group)
	if !ok || or.Operator != LogicalOr || len(or.Items) != 2 {
		t.Fatalf("expr: %#v", filter.Expr)
	}
	if filter.Expr != nil && len(filter.Conditions) != 0 {
		t.Fatal("grouped-only query should not copy items into Conditions")
	}
}

func TestFromURL_GroupFieldDropped(t *testing.T) {
	filter, err := FromURL(url.Values{
		"status":  []string{"active"},
		"or.role": []string{"admin"},
	}, WithAllowedFields("status"))
	if err != nil {
		t.Fatal(err)
	}
	if filter.Expr != nil {
		t.Fatalf("dropped group should leave Expr nil, got %#v", filter.Expr)
	}
	if len(filter.Conditions) != 1 {
		t.Fatalf("conditions: %+v", filter.Conditions)
	}
}

func TestFromURL_GroupValueType(t *testing.T) {
	filter, err := FromURL(url.Values{
		"or.age[gte]": []string{"18", "21"},
	}, WithFieldTypes(map[string]reflect.Type{"age": reflect.TypeOf(int(0))}))
	if err != nil {
		t.Fatal(err)
	}
	or := filter.Expr.(Group)
	if or.Items[0].(Condition).Value != 18 || or.Items[1].(Condition).Value != 21 {
		t.Fatalf("values: %#v", or.Items)
	}
}

func TestFromURL_NoGroupLeavesExprNil(t *testing.T) {
	filter, err := FromURL(url.Values{"status": []string{"active"}})
	if err != nil {
		t.Fatal(err)
	}
	if filter.Expr != nil {
		t.Fatalf("expected nil expr, got %#v", filter.Expr)
	}
}

func TestFromURL_GroupValueInvalid(t *testing.T) {
	_, err := FromURL(url.Values{"or.age": []string{"nope"}}, WithFieldTypes(map[string]reflect.Type{
		"age": reflect.TypeOf(int(0)),
	}))
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseLogicalKey(t *testing.T) {
	if _, _, _, ok := parseLogicalKey("status"); ok {
		t.Fatal("plain key is not logical")
	}
	if _, _, _, ok := parseLogicalKey("or."); ok {
		t.Fatal("empty or key")
	}
	if _, _, _, ok := parseLogicalKey("or.0."); ok {
		t.Fatal("empty field after index")
	}
	op, id, field, ok := parseLogicalKey("and.12.status[ne]")
	if !ok || op != LogicalAnd || id != "12" || field != "status[ne]" {
		t.Fatalf("got %q %q %q %t", op, id, field, ok)
	}
	op, id, field, ok = parseLogicalKey("or.name.extra")
	if !ok || op != LogicalOr || id != "" || field != "name.extra" {
		t.Fatalf("non-index dot: %q %q %q", op, id, field)
	}
}
