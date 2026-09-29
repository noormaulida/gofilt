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

func (o *options) isAllowed(field string) bool {
	if o.allowedFields == nil {
		return true
	}
	return o.allowedFields[field]
}
