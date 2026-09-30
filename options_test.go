package gofilt

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if opts.tagName != "filt" {
		t.Errorf("expected default tagName 'filt', got '%s'", opts.tagName)
	}

	if opts.allowUnknown != false {
		t.Errorf("expected default allowUnknown false, got %t", opts.allowUnknown)
	}

	if opts.allowedFields != nil {
		t.Errorf("expected default allowedFields nil, got %v", opts.allowedFields)
	}
	if opts.allowedSorts != nil {
		t.Errorf("expected default allowedSorts nil, got %v", opts.allowedSorts)
	}

	// Tanpa whitelist, semua field harus diizinkan
	if !opts.isAllowed("any_column") {
		t.Errorf("expected all fields allowed when allowedFields is nil")
	}
}

func TestWithTagName(t *testing.T) {
	opts := defaultOptions()
	WithTagName("custom_tag")(opts)

	if opts.tagName != "custom_tag" {
		t.Errorf("expected tagName 'custom_tag', got '%s'", opts.tagName)
	}

	// Test passing string kosong (harus abaikan dan pakai default/existing)
	WithTagName("")(opts)
	if opts.tagName != "custom_tag" {
		t.Errorf("expected tagName to remain 'custom_tag' on empty input, got '%s'", opts.tagName)
	}
}

func TestWithAllowUnknown(t *testing.T) {
	opts := defaultOptions()
	WithAllowUnknown(true)(opts)

	if !opts.allowUnknown {
		t.Errorf("expected allowUnknown true, got %t", opts.allowUnknown)
	}
}

func TestWithAllowedFields(t *testing.T) {
	opts := defaultOptions()
	WithAllowedFields("name", "age", "status")(opts)

	if opts.allowedFields == nil {
		t.Fatalf("expected allowedFields map to be initialized")
	}

	if len(opts.allowedFields) != 3 {
		t.Errorf("expected 3 allowed fields, got %d", len(opts.allowedFields))
	}

	// Test field verification via isAllowed
	allowedCols := []string{"name", "age", "status"}
	for _, col := range allowedCols {
		if !opts.isAllowed(col) {
			t.Errorf("expected column '%s' to be allowed", col)
		}
	}

	blockedCols := []string{"is_admin", "password", "role"}
	for _, col := range blockedCols {
		if opts.isAllowed(col) {
			t.Errorf("expected column '%s' to be blocked by whitelist", col)
		}
	}
}

func TestWithAllowedFields_EmptyInput(t *testing.T) {
	opts := defaultOptions()
	WithAllowedFields()(opts)

	if opts.allowedFields != nil {
		t.Errorf("expected allowedFields to remain nil on empty slice input")
	}
}

func TestWithAllowedSorts(t *testing.T) {
	opts := defaultOptions()
	WithAllowedSorts("name", "created_at")(opts)

	if !opts.isSortAllowed("name") || !opts.isSortAllowed("created_at") {
		t.Errorf("expected configured sort fields to be allowed")
	}
	if opts.isSortAllowed("is_admin") {
		t.Errorf("expected unknown sort field to be blocked")
	}
}

func TestWithAllowedSorts_EmptyInput(t *testing.T) {
	opts := defaultOptions()
	WithAllowedSorts()(opts)

	if opts.allowedSorts != nil {
		t.Errorf("expected allowedSorts to remain nil on empty input")
	}
	if opts.isSortAllowed("name") {
		t.Errorf("expected sorting to be disabled without a whitelist")
	}
}
