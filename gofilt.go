package gofilt

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

type Condition struct {
	Field    string
	Operator Operator
	Value    any
}

type Filter struct {
	Conditions []Condition
	Sort       []string
	Limit      int
	Offset     int
}

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
