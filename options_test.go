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

func TestWithDefaultLimit(t *testing.T) {
	opts := defaultOptions()
	WithDefaultLimit(25)(opts)
	if opts.defaultLimit != 25 {
		t.Errorf("expected defaultLimit 25, got %d", opts.defaultLimit)
	}

	// Non-positive values should be ignored
	WithDefaultLimit(0)(opts)
	if opts.defaultLimit != 25 {
		t.Errorf("expected defaultLimit to remain 25 after 0, got %d", opts.defaultLimit)
	}

	WithDefaultLimit(-10)(opts)
	if opts.defaultLimit != 25 {
		t.Errorf("expected defaultLimit to remain 25 after negative input, got %d", opts.defaultLimit)
	}
}

func TestWithMaxLimit(t *testing.T) {
	opts := defaultOptions()
	WithMaxLimit(100)(opts)
	if opts.maxLimit != 100 {
		t.Errorf("expected maxLimit 100, got %d", opts.maxLimit)
	}

	// Non-positive values should be ignored
	WithMaxLimit(0)(opts)
	if opts.maxLimit != 100 {
		t.Errorf("expected maxLimit to remain 100 after 0, got %d", opts.maxLimit)
	}

	WithMaxLimit(-50)(opts)
	if opts.maxLimit != 100 {
		t.Errorf("expected maxLimit to remain 100 after negative input, got %d", opts.maxLimit)
	}
}

func TestOptions_ResolveLimit(t *testing.T) {
	opts := defaultOptions()

	// Default unconfigured: requested is used
	if got := opts.resolveLimit(30, true); got != 30 {
		t.Errorf("resolveLimit with requested 30: got %d, want 30", got)
	}
	// Default unconfigured: no requested -> 0
	if got := opts.resolveLimit(0, false); got != 0 {
		t.Errorf("resolveLimit with no requested: got %d, want 0", got)
	}

	// Configured with default limit 20
	opts.defaultLimit = 20
	if got := opts.resolveLimit(0, false); got != 20 {
		t.Errorf("resolveLimit with defaultLimit 20: got %d, want 20", got)
	}
	// Explicit requested overrides default limit
	if got := opts.resolveLimit(50, true); got != 50 {
		t.Errorf("resolveLimit explicit requested 50 overrides default: got %d, want 50", got)
	}

	// Configured with max limit 100
	opts.maxLimit = 100
	if got := opts.resolveLimit(500, true); got != 100 {
		t.Errorf("resolveLimit clamped to maxLimit: got %d, want 100", got)
	}
	if got := opts.resolveLimit(40, true); got != 40 {
		t.Errorf("resolveLimit below maxLimit: got %d, want 40", got)
	}

	// Default limit exceeding max limit is clamped
	opts.defaultLimit = 200
	if got := opts.resolveLimit(0, false); got != 100 {
		t.Errorf("resolveLimit defaultLimit clamped to maxLimit: got %d, want 100", got)
	}
}

