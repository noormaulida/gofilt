package gofilt

func WithTagName(name string) Option {
	return func(o *options) {
		if name != "" {
			o.tagName = name
		}
	}
}

func WithAllowUnknown(allow bool) Option {
	return func(o *options) {
		o.allowUnknown = allow
	}
}

type allowedFieldsOption struct {
	fields map[string]bool
}

func isFieldAllowed(allowed map[string]bool, field string) bool {
	if len(allowed) == 0 {
		return true
	}
	return allowed[field]
}

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
