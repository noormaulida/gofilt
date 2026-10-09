package gofiltsql

import (
	"reflect"
	"strings"
	"testing"

	"github.com/noormaulida/gofilt"
)

func TestBuildWHERE_EmptyFilter(t *testing.T) {
	f := &gofilt.Filter{}

	where, args := BuildWHERE(f)

	if where != "" {
		t.Errorf("expected empty where clause, got %q", where)
	}
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestBuildWHERE_SingleCondition_Eq(t *testing.T) {
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "status", Operator: gofilt.OpEq, Value: "active"},
		},
	}

	where, args := BuildWHERE(f)

	expected := "WHERE status = $1"
	if where != expected {
		t.Errorf("where clause mismatch\ngot:  %s\nwant: %s", where, expected)
	}
	if !reflect.DeepEqual(args, []any{"active"}) {
		t.Errorf("args mismatch: got %v, want [active]", args)
	}
}

func TestBuildWHERE_LikeAndILike(t *testing.T) {
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "name", Operator: gofilt.OpLike, Value: "john"},
			{Field: "email", Operator: gofilt.OpILike, Value: "@gmail"},
		},
	}

	where, args := BuildWHERE(f)

	expected := "WHERE name LIKE $1 AND email ILIKE $2"
	if where != expected {
		t.Errorf("where clause mismatch\ngot:  %s\nwant: %s", where, expected)
	}
	wantArgs := []any{"%john%", "%@gmail%"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args mismatch:\ngot:  %v\nwant: %v", args, wantArgs)
	}
}

func TestBuildWHERE_MultipleOperators(t *testing.T) {
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "age", Operator: gofilt.OpGte, Value: 18},
			{Field: "age", Operator: gofilt.OpLt, Value: 65},
			{Field: "role", Operator: gofilt.OpNe, Value: "guest"},
		},
	}

	where, args := BuildWHERE(f)

	if !strings.HasPrefix(where, "WHERE ") {
		t.Errorf("expected WHERE prefix, got %s", where)
	}

	if strings.Count(where, " AND ") != 2 {
		t.Errorf("expected 2 AND joins for 3 conditions, got clause: %s", where)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if args[0] != 18 || args[1] != 65 || args[2] != "guest" {
		t.Errorf("args mismatch: got %v, want [18 65 guest]", args)
	}

	expectedOrder := []string{"age >= $1", "age < $2", "role != $3"}
	clause := strings.TrimPrefix(where, "WHERE ")
	for _, part := range expectedOrder {
		if !strings.Contains(where, part) {
			t.Errorf("clause missing %q, got: %s", part, clause)
		}
	}
}

func TestBuildWHERE_InOperator_Passthrough(t *testing.T) {
	ids := []string{"1", "2", "3"}
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "id", Operator: gofilt.OpIn, Value: ids},
		},
	}

	where, args := BuildWHERE(f)

	if where != "WHERE id IN $1" {
		t.Errorf("expected 'WHERE id IN $1', got %q", where)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg (the slice), got %d", len(args))
	}
	if !reflect.DeepEqual(args[0], ids) {
		t.Errorf("expected IN value passthrough unchanged, got %v want %v", args[0], ids)
	}
}

func TestBuildWHERE_BetweenOperator_Passthrough(t *testing.T) {
	rangeVal := [2]int{10, 20}
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "price", Operator: gofilt.OpBetween, Value: rangeVal},
		},
	}

	where, args := BuildWHERE(f)

	if where != "WHERE price BETWEEN $1" {
		t.Errorf("expected 'WHERE price BETWEEN $1', got %q", where)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if !reflect.DeepEqual(args[0], rangeVal) {
		t.Errorf("args mismatch: got %v, want %v", args[0], rangeVal)
	}
}

func TestBuildWHERE_NullOperators(t *testing.T) {
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "deleted_at", Operator: gofilt.OpIsNull, Value: "true"},
			{Field: "status", Operator: gofilt.OpEq, Value: "active"},
			{Field: "published_at", Operator: gofilt.OpIsNotNull, Value: true},
		},
	}

	where, args := BuildWHERE(f)
	want := "WHERE deleted_at IS NULL AND status = $1 AND published_at IS NOT NULL"
	if where != want {
		t.Errorf("clause mismatch: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("expected one bound arg [active], got %v", args)
	}
}

func TestBuildORDER(t *testing.T) {
	f := &gofilt.Filter{
		Sorts: []gofilt.Sort{
			{Field: "name", Direction: gofilt.DirectionAsc},
			{Field: "users.created_at", Direction: gofilt.DirectionDesc},
		},
	}

	got := BuildORDER(f)
	want := "ORDER BY name ASC, users.created_at DESC"
	if got != want {
		t.Errorf("order clause mismatch: got %q, want %q", got, want)
	}
}

func TestBuildORDER_SkipsUnsafeSorts(t *testing.T) {
	f := &gofilt.Filter{
		Sorts: []gofilt.Sort{
			{Field: "name; DROP TABLE users", Direction: gofilt.DirectionAsc},
			{Field: "age", Direction: gofilt.Direction("SIDEWAYS")},
		},
	}
	if got := BuildORDER(f); got != "" {
		t.Errorf("expected no clause for invalid sorts, got %q", got)
	}
}

func TestBuildORDER_Empty(t *testing.T) {
	if got := BuildORDER(&gofilt.Filter{}); got != "" {
		t.Errorf("expected empty clause, got %q", got)
	}
}
