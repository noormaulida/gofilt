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

func TestFromURL_UnknownBracketOperatorDefaultsToEq(t *testing.T) {
	queryParams := url.Values{
		"name[wedonthaveop]": []string{"alice"},
	}

	filter, err := FromURL(queryParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(filter.Conditions))
	}
	cond := filter.Conditions[0]
	if cond.Operator != OpEq {
		t.Errorf("unknown bracket op should fall back to OpEq, got %v", cond.Operator)
	}
	if cond.Field != "name" {
		t.Errorf("expected field name, got %s", cond.Field)
	}
	if cond.Value != "alice" {
		t.Errorf("expected value alice, got %v", cond.Value)
	}
}

func TestFromURL_PageAliasForOffset(t *testing.T) {
	queryParams := url.Values{
		"page": []string{"5"},
	}

	filter, err := FromURL(queryParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.Offset != 5 {
		t.Errorf("page alias should set Offset, got Offset=%d", filter.Offset)
	}
}

func TestFromURL_EmptyValuesSkipped(t *testing.T) {
	queryParams := url.Values{
		"name":   []string{""},
		"":       []string{"ghost"},
		"status": []string{},
	}

	filter, err := FromURL(queryParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(filter.Conditions) != 0 {
		t.Errorf("expected no conditions for empty values, got %+v", filter.Conditions)
	}
	if filter.Limit != 0 || filter.Offset != 0 {
		t.Errorf("expected limit/offset zero, got limit=%d offset=%d", filter.Limit, filter.Offset)
	}
}

func TestFromURL_Sorts(t *testing.T) {
	queryParams := url.Values{
		"sort": []string{"name,-created_at", " updated_at "},
	}

	filter, err := FromURL(queryParams,
		WithAllowedSorts("name", "created_at", "updated_at"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []Sort{
		{Field: "name", Direction: DirectionAsc},
		{Field: "created_at", Direction: DirectionDesc},
		{Field: "updated_at", Direction: DirectionAsc},
	}
	if !reflect.DeepEqual(filter.Sorts, want) {
		t.Errorf("sorts mismatch: got %+v, want %+v", filter.Sorts, want)
	}
}

func TestFromURL_SortsRequireWhitelist(t *testing.T) {
	filter, err := FromURL(url.Values{
		"sort": []string{"name,-created_at,, -,unknown"},
	}, WithAllowedSorts("name"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []Sort{{Field: "name", Direction: DirectionAsc}}
	if !reflect.DeepEqual(filter.Sorts, want) {
		t.Errorf("expected only allowed sort, got %+v", filter.Sorts)
	}

	disabled, err := FromURL(url.Values{"sort": []string{"name"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(disabled.Sorts) != 0 {
		t.Errorf("expected sorting disabled without WithAllowedSorts, got %+v", disabled.Sorts)
	}
}

func TestFromURL_WithDefaultLimit(t *testing.T) {
	// When limit parameter is omitted, default limit is applied
	filter, err := FromURL(url.Values{"name": []string{"alice"}}, WithDefaultLimit(20))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", filter.Limit)
	}

	// When limit is explicitly provided, it overrides default limit
	filterWithLimit, err := FromURL(url.Values{"limit": []string{"50"}}, WithDefaultLimit(20))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterWithLimit.Limit != 50 {
		t.Errorf("expected explicit limit 50, got %d", filterWithLimit.Limit)
	}

	// When limit is empty string, default limit is applied
	filterEmptyLimit, err := FromURL(url.Values{"limit": []string{""}}, WithDefaultLimit(20))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterEmptyLimit.Limit != 20 {
		t.Errorf("expected default limit 20 for empty limit param, got %d", filterEmptyLimit.Limit)
	}
}

func TestFromURL_WithMaxLimit(t *testing.T) {
	// Requested limit > maxLimit is clamped to maxLimit
	filter, err := FromURL(url.Values{"limit": []string{"1000"}}, WithMaxLimit(100))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.Limit != 100 {
		t.Errorf("expected limit clamped to 100, got %d", filter.Limit)
	}

	// Requested limit <= maxLimit remains unchanged
	filterUnder, err := FromURL(url.Values{"limit": []string{"50"}}, WithMaxLimit(100))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterUnder.Limit != 50 {
		t.Errorf("expected limit 50, got %d", filterUnder.Limit)
	}

	// Default limit exceeding max limit is clamped
	filterDefaultOverMax, err := FromURL(url.Values{}, WithDefaultLimit(200), WithMaxLimit(100))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterDefaultOverMax.Limit != 100 {
		t.Errorf("expected default limit clamped to max limit 100, got %d", filterDefaultOverMax.Limit)
	}
}

func TestFromURL_NegativeAndInvalidLimit(t *testing.T) {
	// Negative limit returns ErrInvalidLimit
	_, err := FromURL(url.Values{"limit": []string{"-10"}})
	if err != ErrInvalidLimit {
		t.Errorf("expected ErrInvalidLimit for negative limit, got %v", err)
	}

	// Non-numeric limit returns ErrInvalidLimit
	_, err = FromURL(url.Values{"limit": []string{"abc"}})
	if err != ErrInvalidLimit {
		t.Errorf("expected ErrInvalidLimit for non-numeric limit, got %v", err)
	}
}

func TestFromURL_NegativeAndInvalidOffset(t *testing.T) {
	// Negative offset returns ErrInvalidOffset
	_, err := FromURL(url.Values{"offset": []string{"-20"}})
	if err != ErrInvalidOffset {
		t.Errorf("expected ErrInvalidOffset for negative offset, got %v", err)
	}

	// Non-numeric offset returns ErrInvalidOffset
	_, err = FromURL(url.Values{"offset": []string{"xyz"}})
	if err != ErrInvalidOffset {
		t.Errorf("expected ErrInvalidOffset for non-numeric offset, got %v", err)
	}
}

func TestFromURL_NegativeAndInvalidPage(t *testing.T) {
	// Negative page returns ErrInvalidOffset
	_, err := FromURL(url.Values{"page": []string{"-5"}})
	if err != ErrInvalidOffset {
		t.Errorf("expected ErrInvalidOffset for negative page, got %v", err)
	}

	// Non-numeric page returns ErrInvalidOffset
	_, err = FromURL(url.Values{"page": []string{"invalid"}})
	if err != ErrInvalidOffset {
		t.Errorf("expected ErrInvalidOffset for non-numeric page, got %v", err)
	}
}

func TestFromURL_DeterministicErrorOrder(t *testing.T) {
	// When both limit and offset are negative, limit error is returned deterministically
	_, err := FromURL(url.Values{
		"limit":  []string{"-10"},
		"offset": []string{"-20"},
	})
	if err != ErrInvalidLimit {
		t.Errorf("expected ErrInvalidLimit to take precedence, got %v", err)
	}
}

func TestFromURL_OffsetAndPageInteraction(t *testing.T) {
	// When both offset and page are provided and valid, offset takes precedence
	filter, err := FromURL(url.Values{
		"offset": []string{"20"},
		"page":   []string{"5"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.Offset != 20 {
		t.Errorf("expected Offset 20, got %d", filter.Offset)
	}

	// When offset is empty string, page is applied
	filterPageOnly, err := FromURL(url.Values{
		"offset": []string{""},
		"page":   []string{"8"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filterPageOnly.Offset != 8 {
		t.Errorf("expected Offset 8 from page, got %d", filterPageOnly.Offset)
	}

	// When offset is valid but page is invalid, ErrInvalidOffset is returned
	_, err = FromURL(url.Values{
		"offset": []string{"20"},
		"page":   []string{"-1"},
	})
	if err != ErrInvalidOffset {
		t.Errorf("expected ErrInvalidOffset for invalid page alongside valid offset, got %v", err)
	}
}

