package gofilt

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type valueQuery struct {
	Age      int       `filt:"col:age;op:gte"`
	Active   bool      `filt:"col:active"`
	Since    time.Time `filt:"col:created_at;op:gte"`
	Name     string    `filt:"col:name"`
	Skipped  string    `filt:"-"`
	NoTag    string
	Optional *int `filt:"col:score"`
	hidden   int  `filt:"col:hidden"`
}

func TestFieldTypesFrom(t *testing.T) {
	got := FieldTypesFrom(valueQuery{})
	if got["age"] != reflect.TypeOf(int(0)) || got["active"] != reflect.TypeOf(false) || got["created_at"] != reflect.TypeOf(time.Time{}) {
		t.Fatalf("unexpected types: %#v", got)
	}
	if got["name"] != reflect.TypeOf("") {
		t.Fatalf("name type: %v", got["name"])
	}
	if got["score"] != reflect.TypeOf(int(0)) {
		t.Fatalf("pointer field should unwrap to int, got %v", got["score"])
	}
	if _, ok := got["hidden"]; ok {
		t.Fatal("unexported field should be skipped")
	}
	if _, ok := got["Skipped"]; ok {
		t.Fatal("skipped field should be omitted")
	}

	if FieldTypesFrom((*valueQuery)(nil))["age"] != reflect.TypeOf(int(0)) {
		t.Fatal("nil pointer sample should still expose the struct type")
	}
	if FieldTypesFrom(nil) != nil {
		t.Fatal("nil sample should return nil")
	}
	if FieldTypesFrom("nope") != nil {
		t.Fatal("non-struct should return nil")
	}
	if FieldTypesFrom(struct{ N int }{}) != nil {
		t.Fatal("struct without columns should return nil")
	}
}

func TestConvertScalarTypes(t *testing.T) {
	cases := []struct {
		raw  string
		typ  reflect.Type
		want any
	}{
		{"hello", typeString, "hello"},
		{"20", typeInt, int(20)},
		{"-8", typeInt8, int8(-8)},
		{"160", typeInt16, int16(160)},
		{"32000", typeInt32, int32(32000)},
		{"-64", typeInt64, int64(-64)},
		{"7", typeUint, uint(7)},
		{"99", typeUint64, uint64(99)},
		{"1.5", typeFloat32, float32(1.5)},
		{"2.5", typeFloat64, float64(2.5)},
		{"true", typeBool, true},
		{" FALSE ", typeBool, false},
		{"1", typeBool, true},
		{"0", typeBool, false},
	}
	for _, tt := range cases {
		got, err := convertScalar(tt.raw, tt.typ)
		if err != nil {
			t.Fatalf("convertScalar(%q, %s): %v", tt.raw, tt.typ, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("convertScalar(%q, %s) = %#v, want %#v", tt.raw, tt.typ, got, tt.want)
		}
	}

	nano, err := convertScalar("2026-01-02T15:04:05.5Z", typeTime)
	if err != nil {
		t.Fatal(err)
	}
	if nano.(time.Time).UTC().Format(time.RFC3339Nano) != "2026-01-02T15:04:05.5Z" {
		t.Fatalf("nano time: %v", nano)
	}
	clock, err := convertScalar("2026-01-02 15:04:05", typeTime)
	if err != nil {
		t.Fatal(err)
	}
	if clock.(time.Time).Format("2006-01-02 15:04:05") != "2026-01-02 15:04:05" {
		t.Fatalf("clock time: %v", clock)
	}
	day, err := convertScalar("2026-01-02", typeTime)
	if err != nil {
		t.Fatal(err)
	}
	if day.(time.Time).Format("2006-01-02") != "2026-01-02" {
		t.Fatalf("day time: %v", day)
	}
}

func TestConvertScalarInvalid(t *testing.T) {
	bad := []struct {
		raw string
		typ reflect.Type
	}{
		{"x", typeInt},
		{"128", typeInt8},
		{"-1", typeUint},
		{"nope", typeUint64},
		{"x", typeFloat32},
		{"x", typeFloat64},
		{"yes", typeBool},
		{"yesterday", typeTime},
	}
	for _, tt := range bad {
		_, err := convertScalar(tt.raw, tt.typ)
		if !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("convertScalar(%q, %s) err=%v", tt.raw, tt.typ, err)
		}
	}
	if _, err := convertScalar("1", reflect.TypeOf([]byte{})); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("unsupported scalar type err=%v", err)
	}
	if _, err := convertURLValue(1, nil, typeInt); !errors.Is(err, ErrInvalidValue) {
		t.Fatal("non-string value should be invalid")
	}
	if _, err := convertURLValue("1", nil, reflect.TypeOf(uint32(0))); !errors.Is(err, ErrInvalidValue) {
		t.Fatal("unsupported field type should be invalid")
	}
}

func TestConvertSlice(t *testing.T) {
	got, err := convertURLValue(nil, []string{"1", " 2"}, typeInt)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got %#v", got)
	}
	if _, err := convertURLValue(nil, []string{"1", "x"}, typeInt); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("bad element err=%v", err)
	}
}
