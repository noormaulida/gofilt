package gofilt

import (
	"reflect"
	"strings"
	"sync"
)

type tagMeta struct {
	column string
	op     Operator
}

var tagCache sync.Map

// FromStruct builds a Filter by reading the exported fields of input via reflection.
// The input can be a struct value or a pointer to a struct.
//
// Each struct field is scanned for a tag (default name "filt"; override with
// WithTagName). The tag format is `col:<dbcolumn>;op:<operator>`. The operator
// part is optional and defaults to OpEq. Use tag value "-" (or no tag) to skip
// the field.
//
// Fields that hold the zero value for their type are skipped so that unset
// query parameters do not leak into the final Filter.
//
// Example struct:
//
//	type UserQuery struct {
//	    Name   string `filt:"col:name;op:ilike"`
//	    Age    int    `filt:"col:age;op:gte"`
//	    Status string `filt:"col:status"`
//	}
func FromStruct(input any, opts ...Option) (*Filter, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	filter := &Filter{}

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := t.Field(i)

		if fieldVal.IsZero() {
			continue
		}

		meta := parseTag(t, fieldType, cfg.tagName)
		if meta.column == "" {
			continue
		}

		if !cfg.isAllowed(meta.column) {
			continue
		}

		filter.Conditions = append(filter.Conditions, Condition{
			Field:    meta.column,
			Operator: meta.op,
			Value:    fieldVal.Interface(),
		})
	}

	return filter, nil
}

func parseTag(structType reflect.Type, field reflect.StructField, tagName string) tagMeta {
	cacheKey := structType.String() + "." + field.Name
	if cached, ok := tagCache.Load(cacheKey); ok {
		return cached.(tagMeta)
	}

	tag := field.Tag.Get(tagName)
	if tag == "" || tag == "-" {
		meta := tagMeta{}
		tagCache.Store(cacheKey, meta)
		return meta
	}

	meta := tagMeta{op: OpEq}
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "col":
			meta.column = kv[1]
		case "op":
			if op, ok := LookupOperator(kv[1]); ok {
				meta.op = op
			}
		}
	}

	tagCache.Store(cacheKey, meta)
	return meta
}
