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
			meta.op = Operator(strings.ToUpper(kv[1]))
		}
	}

	tagCache.Store(cacheKey, meta)
	return meta
}
