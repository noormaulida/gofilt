package gofilt

// Operator is the comparison operator applied to a filter condition.
// Each operator maps to the equivalent SQL operator when rendered by an adapter.
type Operator string

const (
	OpEq      Operator = "="
	OpNe      Operator = "!="
	OpGt      Operator = ">"
	OpGte     Operator = ">="
	OpLt      Operator = "<"
	OpLte     Operator = "<="
	OpLike    Operator = "LIKE"
	OpILike   Operator = "ILIKE"
	OpIn      Operator = "IN"
	OpBetween Operator = "BETWEEN"
)

// Condition represents a single filter predicate: Field Operator Value.
// Value is typed as any so it can hold scalars (string, int) or slices for OpIn / OpBetween.
type Condition struct {
	Field    string
	Operator Operator
	Value    any
}

// Filter is the output of parsing a struct or URL query.
// It holds an ordered list of conditions plus pagination hints.
// Adapters (GORM, SQL, etc.) accept *Filter and translate it to backend-specific queries.
type Filter struct {
	Conditions []Condition
	Sort       []string
	Limit      int
	Offset     int
}

// Option configures a parser (FromStruct / FromURL).
// Pass options to any parser constructor to override defaults.
type Option func(*options)

type options struct {
	tagName       string
	allowUnknown  bool
	allowedFields map[string]bool
}

func defaultOptions() *options {
	return &options{
		tagName:       "filt",
		allowUnknown:  false,
		allowedFields: nil,
	}
}
