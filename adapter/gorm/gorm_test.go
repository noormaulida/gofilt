package gofiltgorm

import (
	"strings"
	"testing"

	"github.com/noormaulida/gofilt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testUser struct {
	ID     int
	Name   string
	Age    int
	Status string
}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory: %v", err)
	}
	err = db.AutoMigrate(&testUser{})
	if err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	seed := []testUser{
		{ID: 1, Name: "Alice", Age: 25, Status: "active"},
		{ID: 2, Name: "Bob", Age: 30, Status: "active"},
		{ID: 3, Name: "Charlie", Age: 17, Status: "pending"},
		{ID: 4, Name: "Diana", Age: 40, Status: "archived"},
		{ID: 5, Name: "Eve", Age: 50, Status: "active"},
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	return db
}

func dryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("failed open dry db: %v", err)
	}
	return db
}

func ids(users []testUser) []int {
	out := make([]int, len(users))
	for i, u := range users {
		out[i] = u.ID
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestApply_EqOperator(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "status", Operator: gofilt.OpEq, Value: "pending"},
		},
	}

	var results []testUser
	err := Apply(db, f).Order("id").Find(&results).Error
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	want := []int{3}
	if got := ids(results); !equalInts(got, want) {
		t.Errorf("Eq filter — ids: got %v, want %v", got, want)
	}
}

func TestApply_NeOperator(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "status", Operator: gofilt.OpNe, Value: "active"},
		},
	}

	var results []testUser
	err := Apply(db, f).Order("id").Find(&results).Error
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	want := []int{3, 4}
	if got := ids(results); !equalInts(got, want) {
		t.Errorf("Ne filter — ids: got %v, want %v", got, want)
	}
}

func TestApply_GtGteLtLte(t *testing.T) {
	db := openDB(t)

	t.Run("gte", func(t *testing.T) {
		f := &gofilt.Filter{
			Conditions: []gofilt.Condition{
				{Field: "age", Operator: gofilt.OpGte, Value: 40},
			},
		}
		var results []testUser
		if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(results), []int{4, 5}; !equalInts(got, want) {
			t.Errorf("gte: got %v want %v", got, want)
		}
	})

	t.Run("gt", func(t *testing.T) {
		f := &gofilt.Filter{
			Conditions: []gofilt.Condition{
				{Field: "age", Operator: gofilt.OpGt, Value: 40},
			},
		}
		var results []testUser
		if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(results), []int{5}; !equalInts(got, want) {
			t.Errorf("gt: got %v want %v", got, want)
		}
	})

	t.Run("lte", func(t *testing.T) {
		f := &gofilt.Filter{
			Conditions: []gofilt.Condition{
				{Field: "age", Operator: gofilt.OpLte, Value: 25},
			},
		}
		var results []testUser
		if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(results), []int{1, 3}; !equalInts(got, want) {
			t.Errorf("lte: got %v want %v", got, want)
		}
	})

	t.Run("lt", func(t *testing.T) {
		f := &gofilt.Filter{
			Conditions: []gofilt.Condition{
				{Field: "age", Operator: gofilt.OpLt, Value: 25},
			},
		}
		var results []testUser
		if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(results), []int{3}; !equalInts(got, want) {
			t.Errorf("lt: got %v want %v", got, want)
		}
	})
}

func TestApply_Like(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "name", Operator: gofilt.OpLike, Value: "li"},
		},
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}
	if got, want := ids(results), []int{1, 3}; !equalInts(got, want) {
		t.Errorf("like 'li': got %v want [1 3] (Alice+Charlie both contain 'li')", got)
	}
}

func TestApply_ILike_DryRun(t *testing.T) {
	db := dryDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "name", Operator: gofilt.OpILike, Value: "cHar"},
		},
	}

	tx := Apply(db, f).Model(&testUser{}).Find(&[]testUser{})
	sql := tx.Statement.SQL.String()
	vars := tx.Statement.Vars

	if !strings.Contains(sql, "ILIKE") {
		t.Errorf("expected ILIKE in rendered SQL, got: %s", sql)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 stmt var, got %d (vars=%v)", len(vars), vars)
	}
	want := "%cHar%"
	if got, _ := vars[0].(string); got != want {
		t.Errorf("ILike value not wrapped: got %q, want %q", got, want)
	}
}

func TestApply_splitBetween_Nil(t *testing.T) {
	lo, hi, ok := splitBetween(nil)
	if ok {
		t.Errorf("splitBetween(nil) should return ok=false, got ok=true lo=%v hi=%v", lo, hi)
	}
	lo2, hi2, ok2 := splitBetween("string")
	if ok2 {
		t.Errorf("splitBetween(string) should return ok=false, got ok=true lo=%v hi=%v", lo2, hi2)
	}
	shortSlice := []int{1}
	lo3, hi3, ok3 := splitBetween(shortSlice)
	if ok3 {
		t.Errorf("splitBetween(len=1 slice) should ok=false, got ok=true lo=%v hi=%v", lo3, hi3)
	}
	longSlice := []int{1, 2, 3}
	lo4, hi4, ok4 := splitBetween(longSlice)
	if ok4 {
		t.Errorf("splitBetween(len=3 slice) should ok=false, got ok=true lo=%v hi=%v", lo4, hi4)
	}
}

func TestApply_InOperator(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "id", Operator: gofilt.OpIn, Value: []int{1, 4, 5}},
		},
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if got, want := ids(results), []int{1, 4, 5}; !equalInts(got, want) {
		t.Errorf("IN: got %v want %v", got, want)
	}
}

func TestApply_BetweenOperator_Slice(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "age", Operator: gofilt.OpBetween, Value: []int{25, 40}},
		},
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if got, want := ids(results), []int{1, 2, 4}; !equalInts(got, want) {
		t.Errorf("BETWEEN [25,40]: got %v want %v", got, want)
	}
}

func TestApply_BetweenOperator_Array(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "age", Operator: gofilt.OpBetween, Value: [2]int{17, 30}},
		},
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if got, want := ids(results), []int{1, 2, 3}; !equalInts(got, want) {
		t.Errorf("BETWEEN array [17,30]: got %v want %v", got, want)
	}
}

func TestApply_BetweenOperator_NilFallback(t *testing.T) {
	db := dryDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "age", Operator: gofilt.OpBetween, Value: "not a slice"},
		},
	}

	tx := Apply(db, f).Model(&testUser{}).Find(&[]testUser{})
	sql := tx.Statement.SQL.String()

	if !strings.Contains(sql, "BETWEEN") {
		t.Errorf("expected BETWEEN keyword, got: %s", sql)
	}
}

func TestApply_ExpressionTree(t *testing.T) {
	db := dryDB(t)
	f := &gofilt.Filter{Expr: &gofilt.Group{Operator: gofilt.LogicalAnd, Items: []gofilt.Expression{
		&gofilt.Condition{Field: "status", Operator: gofilt.OpEq, Value: "active"},
		gofilt.Or(
			gofilt.Condition{Field: "role", Operator: gofilt.OpEq, Value: "admin"},
			gofilt.Condition{Field: "role", Operator: gofilt.OpEq, Value: "moderator"},
		),
		(*gofilt.Condition)(nil),
		nil,
		gofilt.Group{},
		(*gofilt.Group)(nil),
	}}}

	tx := Apply(db, f).Model(&testUser{}).Find(&[]testUser{})
	sql := tx.Statement.SQL.String()
	if !strings.Contains(sql, "status = ?") || !strings.Contains(sql, "(role = ? OR role = ?)") {
		t.Errorf("expression sql: %s", sql)
	}
	if strings.Contains(sql, "ignored") {
		t.Errorf("expr should replace conditions, got %s", sql)
	}

	empty := Apply(dryDB(t), &gofilt.Filter{Expr: gofilt.Group{}}).Model(&testUser{}).Find(&[]testUser{})
	if strings.Contains(empty.Statement.SQL.String(), "WHERE") {
		t.Errorf("empty group should not add WHERE, got %s", empty.Statement.SQL.String())
	}
}

func TestApply_NullOperators(t *testing.T) {
	db := dryDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "deleted_at", Operator: gofilt.OpIsNull, Value: true},
			{Field: "published_at", Operator: gofilt.OpIsNotNull, Value: true},
		},
	}

	tx := Apply(db, f).Model(&testUser{}).Find(&[]testUser{})
	sql := tx.Statement.SQL.String()
	if !strings.Contains(sql, "deleted_at IS NULL") {
		t.Errorf("expected IS NULL clause, got: %s", sql)
	}
	if !strings.Contains(sql, "published_at IS NOT NULL") {
		t.Errorf("expected IS NOT NULL clause, got: %s", sql)
	}
	if strings.Contains(sql, "?") {
		t.Errorf("null operators should not bind a value, got: %s", sql)
	}
}

func TestApply_LimitAndOffset(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "status", Operator: gofilt.OpEq, Value: "active"},
		},
		Limit:  1,
		Offset: 1,
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if got, want := ids(results), []int{2}; !equalInts(got, want) {
		t.Errorf("limit/offset active users: got %v want [2] (Bob, the 2nd active after Alice)", got)
	}
}

func TestApply_NoLimitZeroNotApplied(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{Limit: 0, Offset: 0}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if len(results) != 5 {
		t.Errorf("expected all 5 rows with Limit=0 Offset=0, got %d", len(results))
	}
}

func TestApply_EmptyConditions(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if len(results) != 5 {
		t.Errorf("expected all 5 rows when no conditions, got %d", len(results))
	}
}

func TestApply_MultipleConditions(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Conditions: []gofilt.Condition{
			{Field: "status", Operator: gofilt.OpEq, Value: "active"},
			{Field: "age", Operator: gofilt.OpGt, Value: 25},
		},
	}

	var results []testUser
	if err := Apply(db, f).Order("id").Find(&results).Error; err != nil {
		t.Fatal(err)
	}

	if got, want := ids(results), []int{2, 5}; !equalInts(got, want) {
		t.Errorf("AND multi-cond: got %v want %v (active AND age>25)", got, want)
	}
}

func TestApply_Sorts(t *testing.T) {
	db := openDB(t)
	f := &gofilt.Filter{
		Sorts: []gofilt.Sort{
			{Field: "age", Direction: gofilt.DirectionDesc},
			{Field: "name", Direction: gofilt.DirectionAsc},
		},
	}

	var results []testUser
	if err := Apply(db, f).Find(&results).Error; err != nil {
		t.Fatal(err)
	}
	if got, want := ids(results), []int{5, 4, 2, 1, 3}; !equalInts(got, want) {
		t.Errorf("sorted ids: got %v want %v", got, want)
	}
}

func TestApply_SkipsUnsafeSorts(t *testing.T) {
	db := dryDB(t)
	f := &gofilt.Filter{
		Sorts: []gofilt.Sort{
			{Field: "name; DROP TABLE users", Direction: gofilt.DirectionAsc},
			{Field: "age", Direction: gofilt.Direction("SIDEWAYS")},
			{Field: "users.name", Direction: gofilt.DirectionAsc},
		},
	}

	tx := Apply(db, f).Model(&testUser{}).Find(&[]testUser{})
	sql := tx.Statement.SQL.String()
	if strings.Contains(sql, "DROP TABLE") || strings.Contains(sql, "SIDEWAYS") {
		t.Errorf("unsafe sort reached SQL: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY users.name ASC") {
		t.Errorf("expected valid qualified sort in SQL, got: %s", sql)
	}
}
