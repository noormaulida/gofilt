package gofilt

import (
	"errors"
	"reflect"
	"regexp"
)

var (
	// ErrInvalidLimit is returned when a limit parameter cannot be parsed or is negative.
	ErrInvalidLimit = errors.New("gofilt: invalid limit")

	// ErrInvalidOffset is returned when an offset or page parameter cannot be parsed or is negative.
	ErrInvalidOffset = errors.New("gofilt: invalid offset")

	// ErrInvalidValue is returned when a URL value cannot be converted to the
	// type registered for its field with WithFieldTypes.
	ErrInvalidValue = errors.New("gofilt: invalid value")
)

// Operator is the comparison operator applied to a filter condition.
// Each operator maps to the equivalent SQL operator when rendered by an adapter.
type Operator string

const (
	OpEq        Operator = "="
	OpNe        Operator = "!="
	OpGt        Operator = ">"
	OpGte       Operator = ">="
	OpLt        Operator = "<"
	OpLte       Operator = "<="
	OpLike      Operator = "LIKE"
	OpILike     Operator = "ILIKE"
	OpIn        Operator = "IN"
	OpBetween   Operator = "BETWEEN"
	OpIsNull    Operator = "IS NULL"
	OpIsNotNull Operator = "IS NOT NULL"
)

// Condition represents a single filter predicate: Field Operator Value.
// Value is typed as any so it can hold scalars (string, int) or slices for OpIn / OpBetween.
// Condition implements Expression, so it can sit inside a Group.
type Condition struct {
	Field    string
	Operator Operator
	Value    any
}

func (Condition) expression() {}

// LogicalOperator joins the items of a Group.
type LogicalOperator string

const (
	// LogicalAnd requires every item in the group to match.
	LogicalAnd LogicalOperator = "AND"
	// LogicalOr requires at least one item in the group to match.
	LogicalOr LogicalOperator = "OR"
)

// Expression is a node in a filter tree. The concrete types are Condition and Group.
type Expression interface {
	expression()
}

// Group joins nested expressions with AND or OR.
// Items may be Condition values, Group values, or pointers to either.
// A LogicalOperator other than LogicalOr is rendered as AND.
type Group struct {
	Operator LogicalOperator
	Items    []Expression
}

func (Group) expression() {}

// And builds a group that requires every item to match.
func And(items ...Expression) Group {
	return Group{Operator: LogicalAnd, Items: items}
}

// Or builds a group that requires any item to match.
func Or(items ...Expression) Group {
	return Group{Operator: LogicalOr, Items: items}
}

// Direction is the order applied to a sort field.
type Direction string

const (
	// DirectionAsc sorts values in ascending order.
	DirectionAsc Direction = "ASC"
	// DirectionDesc sorts values in descending order.
	DirectionDesc Direction = "DESC"
)

// Sort represents one validated sort field and its direction.
type Sort struct {
	Field     string
	Direction Direction
}

// Filter is the output of parsing a struct or URL query.
// It holds ordered conditions, an optional expression tree, sorts, and pagination hints.
// Adapters (GORM, SQL, etc.) accept *Filter and translate it to backend-specific queries.
//
// Conditions is an implicit AND list. Expr, when set, is the tree adapters render
// instead, which is how nested AND and OR groups are expressed.
//
// Search is a free-text term matched with ILIKE against SearchFields. Both stay
// empty unless FromURL is called with WithSearchFields. Fields is an ordered
// column projection from the fields query parameter, populated only when
// WithSelectFields allows those names.
type Filter struct {
	Conditions   []Condition
	Expr         Expression
	Search       string
	SearchFields []string
	Fields       []string
	Sorts        []Sort
	Limit        int
	Offset       int
}

// columnIdent matches a single SQL identifier or a dotted path such as users.name.
// It rejects spaces, quotes, and statement separators.
var columnIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// safeColumns returns names that match columnIdent, in order, without duplicates.
func safeColumns(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		if name == "" || seen[name] || !columnIdent.MatchString(name) {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Option configures a parser (FromStruct / FromURL).
// Pass options to any parser constructor to override defaults.
type Option func(*options)

type options struct {
	tagName       string
	allowUnknown  bool
	allowedFields map[string]bool
	allowedSorts  map[string]bool
	defaultLimit  int
	maxLimit      int
	fieldTypes    map[string]reflect.Type
	searchFields  []string
	selectFields  map[string]bool
}

func defaultOptions() *options {
	return &options{
		tagName:       "filt",
		allowUnknown:  false,
		allowedFields: nil,
		allowedSorts:  nil,
		defaultLimit:  0,
		maxLimit:      0,
	}
}
