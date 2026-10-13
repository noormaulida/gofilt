package gofilt

import (
	"net/url"
	"reflect"
	"testing"
)

func TestFromURL_SearchRequiresExplicitFields(t *testing.T) {
	plain, err := FromURL(url.Values{"q": []string{"noor"}})
	if err != nil {
		t.Fatal(err)
	}
	if plain.Search != "" || len(plain.Conditions) != 1 || plain.Conditions[0].Field != "q" {
		t.Fatalf("q should stay a normal column, got search=%q conditions=%+v", plain.Search, plain.Conditions)
	}

	filter, err := FromURL(url.Values{
		"q":      []string{"noor"},
		"search": []string{"other"},
		"status": []string{"active"},
	}, WithSearchFields("name", "email", "phone", "name", "bad name", ""))
	if err != nil {
		t.Fatal(err)
	}
	if filter.Search != "noor" {
		t.Fatalf("search term: %q", filter.Search)
	}
	if !reflect.DeepEqual(filter.SearchFields, []string{"name", "email", "phone"}) {
		t.Fatalf("search fields: %#v", filter.SearchFields)
	}
	if len(filter.Conditions) != 1 || filter.Conditions[0].Field != "status" {
		t.Fatalf("q must not also be a condition: %+v", filter.Conditions)
	}
}

func TestFromURL_SearchAlias(t *testing.T) {
	filter, err := FromURL(url.Values{"search": []string{"", "noor"}}, WithSearchFields("email"))
	if err != nil {
		t.Fatal(err)
	}
	if filter.Search != "" {
		t.Fatalf("only the first search value is read, got %q", filter.Search)
	}

	filter, err = FromURL(url.Values{"search": []string{"noor"}}, WithSearchFields("users.email"))
	if err != nil {
		t.Fatal(err)
	}
	if filter.Search != "noor" || !reflect.DeepEqual(filter.SearchFields, []string{"users.email"}) {
		t.Fatalf("alias search: %#v", filter)
	}

	empty, err := FromURL(url.Values{"status": []string{"active"}}, WithSearchFields("name"))
	if err != nil {
		t.Fatal(err)
	}
	if empty.Search != "" || empty.SearchFields != nil {
		t.Fatalf("missing term should not enable search: %#v", empty)
	}
}

func TestFromURL_SelectFields(t *testing.T) {
	plain, err := FromURL(url.Values{"fields": []string{"id,name"}})
	if err != nil {
		t.Fatal(err)
	}
	if plain.Fields != nil || len(plain.Conditions) != 1 || plain.Conditions[0].Operator != OpIn {
		t.Fatalf("fields should stay a normal column without an allowlist: %+v", plain)
	}

	filter, err := FromURL(url.Values{"fields": []string{" id, name, email, name, ,drop"}}, WithSelectFields("id", "name", "email", "id"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(filter.Fields, []string{"id", "name", "email"}) {
		t.Fatalf("fields: %#v", filter.Fields)
	}
	if len(filter.Conditions) != 0 {
		t.Fatalf("fields must not also be a condition: %+v", filter.Conditions)
	}

	none, err := FromURL(url.Values{"status": []string{"active"}}, WithSelectFields("id"))
	if err != nil {
		t.Fatal(err)
	}
	if none.Fields != nil {
		t.Fatalf("missing fields param: %#v", none.Fields)
	}
}

func TestSearchAndSelectOptionsNoop(t *testing.T) {
	opts := defaultOptions()
	WithSearchFields()(opts)
	WithSearchFields("a b")(opts)
	if opts.searchFields != nil {
		t.Fatalf("unsafe or empty search fields should be a no-op: %#v", opts.searchFields)
	}
	WithSelectFields()(opts)
	WithSelectFields("a;b")(opts)
	if opts.selectFields != nil {
		t.Fatalf("unsafe or empty select fields should be a no-op: %#v", opts.selectFields)
	}
}

func TestSafeColumns(t *testing.T) {
	if safeColumns(nil) != nil {
		t.Fatal("nil input")
	}
	if safeColumns([]string{"", "a b"}) != nil {
		t.Fatal("only unsafe names should return nil")
	}
}
