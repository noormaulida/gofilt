package gofilt

import "math"

// Pagination holds pagination parameters and provides helper methods
// to compute page numbers and adjacent offsets.
//
// Total record count is intentionally excluded because determining total records
// requires a COUNT(*) database query, which is the responsibility of the repository layer.
type Pagination struct {
	Limit  int
	Offset int
}

// NewPagination creates a new Pagination instance with the given limit and offset.
func NewPagination(limit, offset int) Pagination {
	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

// Pagination returns the Pagination helper for the Filter.
// If f is nil, a zero-valued Pagination is returned.
func (f *Filter) Pagination() Pagination {
	if f == nil {
		return Pagination{}
	}
	return Pagination{
		Limit:  f.Limit,
		Offset: f.Offset,
	}
}

// Page returns the 1-based page number computed from Offset and Limit.
// If Limit <= 0 or Offset <= 0, it returns 1.
func (p Pagination) Page() int {
	if p.Limit <= 0 || p.Offset <= 0 {
		return 1
	}
	page := (p.Offset / p.Limit) + 1
	if page < 1 {
		return math.MaxInt
	}
	return page
}

// NextOffset returns the offset for the next page.
// If Limit <= 0, it returns the current offset (or 0 if negative).
// If an integer overflow would occur, math.MaxInt is returned.
func (p Pagination) NextOffset() int {
	baseOffset := p.Offset
	if baseOffset < 0 {
		baseOffset = 0
	}
	if p.Limit <= 0 {
		return baseOffset
	}
	if math.MaxInt-p.Limit < baseOffset {
		return math.MaxInt
	}
	return baseOffset + p.Limit
}

// PreviousOffset returns the offset for the previous page, clamped to 0.
// If Limit <= 0 or Offset <= 0, it returns 0.
func (p Pagination) PreviousOffset() int {
	if p.Limit <= 0 || p.Offset <= 0 {
		return 0
	}
	prev := p.Offset - p.Limit
	if prev < 0 {
		return 0
	}
	return prev
}
