package gofilt

import (
	"math"
	"testing"
)

func TestNewPagination(t *testing.T) {
	p := NewPagination(20, 40)
	if p.Limit != 20 {
		t.Errorf("expected limit 20, got %d", p.Limit)
	}
	if p.Offset != 40 {
		t.Errorf("expected offset 40, got %d", p.Offset)
	}
}

func TestFilter_Pagination(t *testing.T) {
	f := &Filter{
		Limit:  15,
		Offset: 30,
	}

	p := f.Pagination()
	if p.Limit != 15 || p.Offset != 30 {
		t.Errorf("expected limit 15 and offset 30, got %+v", p)
	}

	var nilFilter *Filter
	nilP := nilFilter.Pagination()
	if nilP.Limit != 0 || nilP.Offset != 0 {
		t.Errorf("expected zero-value pagination on nil filter, got %+v", nilP)
	}
}

func TestPagination_Page(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected int
	}{
		{
			name:     "first page, exact offset 0",
			p:        Pagination{Limit: 10, Offset: 0},
			expected: 1,
		},
		{
			name:     "second page, exact offset 10",
			p:        Pagination{Limit: 10, Offset: 10},
			expected: 2,
		},
		{
			name:     "third page, exact offset 20",
			p:        Pagination{Limit: 10, Offset: 20},
			expected: 3,
		},
		{
			name:     "second page, partial offset 15",
			p:        Pagination{Limit: 10, Offset: 15},
			expected: 2,
		},
		{
			name:     "first page, offset smaller than limit",
			p:        Pagination{Limit: 10, Offset: 5},
			expected: 1,
		},
		{
			name:     "zero limit returns page 1",
			p:        Pagination{Limit: 0, Offset: 20},
			expected: 1,
		},
		{
			name:     "negative limit returns page 1",
			p:        Pagination{Limit: -5, Offset: 20},
			expected: 1,
		},
		{
			name:     "negative offset returns page 1",
			p:        Pagination{Limit: 10, Offset: -10},
			expected: 1,
		},
		{
			name:     "overflow offset returns math.MaxInt",
			p:        Pagination{Limit: 1, Offset: math.MaxInt},
			expected: math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Page()
			if got != tt.expected {
				t.Errorf("Page() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPagination_NextOffset(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected int
	}{
		{
			name:     "first page to second page",
			p:        Pagination{Limit: 10, Offset: 0},
			expected: 10,
		},
		{
			name:     "second page to third page",
			p:        Pagination{Limit: 10, Offset: 10},
			expected: 20,
		},
		{
			name:     "zero limit returns current offset",
			p:        Pagination{Limit: 0, Offset: 15},
			expected: 15,
		},
		{
			name:     "zero limit with negative offset returns 0",
			p:        Pagination{Limit: 0, Offset: -5},
			expected: 0,
		},
		{
			name:     "negative limit returns current offset",
			p:        Pagination{Limit: -10, Offset: 20},
			expected: 20,
		},
		{
			name:     "negative offset clamped to 0",
			p:        Pagination{Limit: 10, Offset: -5},
			expected: 10,
		},
		{
			name:     "overflow returns math.MaxInt",
			p:        Pagination{Limit: 10, Offset: math.MaxInt - 5},
			expected: math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.NextOffset()
			if got != tt.expected {
				t.Errorf("NextOffset() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPagination_PreviousOffset(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected int
	}{
		{
			name:     "third page to second page",
			p:        Pagination{Limit: 10, Offset: 20},
			expected: 10,
		},
		{
			name:     "second page to first page",
			p:        Pagination{Limit: 10, Offset: 10},
			expected: 0,
		},
		{
			name:     "partial offset clamped to 0",
			p:        Pagination{Limit: 10, Offset: 5},
			expected: 0,
		},
		{
			name:     "first page returns 0",
			p:        Pagination{Limit: 10, Offset: 0},
			expected: 0,
		},
		{
			name:     "negative offset returns 0",
			p:        Pagination{Limit: 10, Offset: -10},
			expected: 0,
		},
		{
			name:     "zero limit returns 0",
			p:        Pagination{Limit: 0, Offset: 20},
			expected: 0,
		},
		{
			name:     "negative limit returns 0",
			p:        Pagination{Limit: -5, Offset: 20},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.PreviousOffset()
			if got != tt.expected {
				t.Errorf("PreviousOffset() = %d, want %d", got, tt.expected)
			}
		})
	}
}
