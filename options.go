package gofilt

// WithTagName overrides the default struct tag name used by FromStruct.
// The default tag name is "filt". Passing an empty string is a no-op and
// the previous (or default) value is kept.
func WithTagName(name string) Option {
	return func(o *options) {
		if name != "" {
			o.tagName = name
		}
	}
}

// WithAllowUnknown controls whether unrecognized input is silently dropped
// or rejected. Currently reserved for future URL-parser strict mode; using
// this option has no observable effect in the current version.
func WithAllowUnknown(allow bool) Option {
	return func(o *options) {
		o.allowUnknown = allow
	}
}

type allowedFieldsOption struct {
	fields map[string]bool
}

// WithAllowedFields restricts the set of columns parsers are allowed to emit.
// Any parsed field whose column name is not in the whitelist is dropped
// from the output Filter. Call with no arguments is a no-op.
//
// When this option is not set (the default), all fields present in the input
// are passed through.
func WithAllowedFields(fields ...string) Option {
	return func(o *options) {
		if len(fields) == 0 {
			return
		}
		allowed := make(map[string]bool, len(fields))
		for _, f := range fields {
			allowed[f] = true
		}
		o.allowedFields = allowed
	}
}

// WithAllowedSorts restricts the fields that FromURL may use for sorting.
// Sort parameters are ignored unless this option is configured. Unknown sort
// fields are silently dropped. Call with no arguments is a no-op.
func WithAllowedSorts(fields ...string) Option {
	return func(o *options) {
		if len(fields) == 0 {
			return
		}
		allowed := make(map[string]bool, len(fields))
		for _, f := range fields {
			allowed[f] = true
		}
		o.allowedSorts = allowed
	}
}

// WithDefaultLimit sets a fallback limit to use when no limit parameter is
// specified in the input. If limit <= 0, this option is ignored.
func WithDefaultLimit(limit int) Option {
	return func(o *options) {
		if limit > 0 {
			o.defaultLimit = limit
		}
	}
}

// WithMaxLimit sets an upper bound on the allowed limit. If a requested limit
// exceeds max, it is clamped to max. If max <= 0, this option is ignored.
func WithMaxLimit(max int) Option {
	return func(o *options) {
		if max > 0 {
			o.maxLimit = max
		}
	}
}

func (o *options) resolveLimit(requested int, hasRequested bool) int {
	var limit int
	if hasRequested {
		limit = requested
	} else if o.defaultLimit > 0 {
		limit = o.defaultLimit
	}
	if o.maxLimit > 0 && limit > o.maxLimit {
		limit = o.maxLimit
	}
	return limit
}

func (o *options) isAllowed(field string) bool {
	if o.allowedFields == nil {
		return true
	}
	return o.allowedFields[field]
}

func (o *options) isSortAllowed(field string) bool {
	return o.allowedSorts != nil && o.allowedSorts[field]
}
