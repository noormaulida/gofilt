package gofilt

import (
	"net/url"
	"reflect"
	"testing"
)

func TestFromURL_OperatorsAndPagination(t *testing.T) {
	queryParams := url.Values{
		"name[ilike]": []string{"john"},
		"age[gte]":    []string{"21"},
		"limit":       []string{"10"},
		"offset":      []string{"20"},
	}

	filter, err := FromURL(queryParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if filter.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", filter.Limit)
	}
	if filter.Offset != 20 {
		t.Errorf("expected Offset 20, got %d", filter.Offset)
	}

	if len(filter.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(filter.Conditions))
	}
}

func TestFromURL_InOperatorAndCommas(t *testing.T) {
	queryParams := url.Values{
		"status": []string{"active,pending,archived"},
	}

	filter, err := FromURL(queryParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(filter.Conditions))
	}

	cond := filter.Conditions[0]
	if cond.Operator != OpIn {
		t.Errorf("expected operator OpIn, got %v", cond.Operator)
	}

	expectedSlice := []string{"active", "pending", "archived"}
	if !reflect.DeepEqual(cond.Value, expectedSlice) {
		t.Errorf("expected slice %v, got %v", expectedSlice, cond.Value)
	}
}

func TestFromURL_WithAllowedFields(t *testing.T) {
	queryParams := url.Values{
		"name[like]": []string{"john"},
		"is_admin":   []string{"true"}, // Malicious query param
	}

	filter, err := FromURL(queryParams, WithAllowedFields("name"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition after whitelist, got %d", len(filter.Conditions))
	}

	if filter.Conditions[0].Field != "name" {
		t.Errorf("expected field 'name', got %s", filter.Conditions[0].Field)
	}
}
