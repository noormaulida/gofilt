package gofilt

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	typeString  = reflect.TypeOf("")
	typeInt     = reflect.TypeOf(int(0))
	typeInt8    = reflect.TypeOf(int8(0))
	typeInt16   = reflect.TypeOf(int16(0))
	typeInt32   = reflect.TypeOf(int32(0))
	typeInt64   = reflect.TypeOf(int64(0))
	typeUint    = reflect.TypeOf(uint(0))
	typeUint64  = reflect.TypeOf(uint64(0))
	typeFloat32 = reflect.TypeOf(float32(0))
	typeFloat64 = reflect.TypeOf(float64(0))
	typeBool    = reflect.TypeOf(false)
	typeTime    = reflect.TypeOf(time.Time{})
)

var supportedFieldTypes = map[reflect.Type]struct{}{
	typeString:  {},
	typeInt:     {},
	typeInt8:    {},
	typeInt16:   {},
	typeInt32:   {},
	typeInt64:   {},
	typeUint:    {},
	typeUint64:  {},
	typeFloat32: {},
	typeFloat64: {},
	typeBool:    {},
	typeTime:    {},
}

// FieldTypesFrom reads column types from a query struct (or a pointer to one).
// Column names come from the default "filt" tag (`col:<name>`). Fields without
// a column, unexported fields, and fields tagged "-" are skipped. Pointer
// fields contribute the type they point to.
//
// The returned map is ready to pass to WithFieldTypes. A nil or non-struct
// sample returns nil.
//
//	type UserQuery struct {
//	    Age    int       `filt:"col:age;op:gte"`
//	    Active bool      `filt:"col:active"`
//	    Since  time.Time `filt:"col:created_at;op:gte"`
//	}
//	types := gofilt.FieldTypesFrom(UserQuery{})
func FieldTypesFrom(sample any) map[string]reflect.Type {
	if sample == nil {
		return nil
	}
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	out := make(map[string]reflect.Type)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		meta := parseTag(t, field, "filt")
		if meta.column == "" {
			continue
		}
		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		out[meta.column] = ft
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func convertURLValue(parsed any, parts []string, typ reflect.Type) (any, error) {
	if _, ok := supportedFieldTypes[typ]; !ok {
		return nil, ErrInvalidValue
	}
	if parts != nil {
		return convertSlice(parts, typ)
	}
	raw, ok := parsed.(string)
	if !ok {
		return nil, ErrInvalidValue
	}
	return convertScalar(raw, typ)
}

func convertSlice(parts []string, typ reflect.Type) (any, error) {
	slice := reflect.MakeSlice(reflect.SliceOf(typ), 0, len(parts))
	for _, part := range parts {
		v, err := convertScalar(part, typ)
		if err != nil {
			return nil, err
		}
		slice = reflect.Append(slice, reflect.ValueOf(v))
	}
	return slice.Interface(), nil
}

func convertScalar(raw string, typ reflect.Type) (any, error) {
	switch typ {
	case typeString:
		return raw, nil
	case typeInt:
		return parseSigned(raw, 0, func(n int64) any { return int(n) })
	case typeInt8:
		return parseSigned(raw, 8, func(n int64) any { return int8(n) })
	case typeInt16:
		return parseSigned(raw, 16, func(n int64) any { return int16(n) })
	case typeInt32:
		return parseSigned(raw, 32, func(n int64) any { return int32(n) })
	case typeInt64:
		return parseSigned(raw, 64, func(n int64) any { return n })
	case typeUint:
		return parseUnsigned(raw, 0, func(n uint64) any { return uint(n) })
	case typeUint64:
		return parseUnsigned(raw, 64, func(n uint64) any { return n })
	case typeFloat32:
		return parseFloat(raw, 32, func(n float64) any { return float32(n) })
	case typeFloat64:
		return parseFloat(raw, 64, func(n float64) any { return n })
	case typeBool:
		return parseBool(raw)
	case typeTime:
		return parseTime(raw)
	default:
		return nil, ErrInvalidValue
	}
}

func parseSigned(raw string, bitSize int, box func(int64) any) (any, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, bitSize)
	if err != nil {
		return nil, ErrInvalidValue
	}
	return box(n), nil
}

func parseUnsigned(raw string, bitSize int, box func(uint64) any) (any, error) {
	n, err := strconv.ParseUint(strings.TrimSpace(raw), 10, bitSize)
	if err != nil {
		return nil, ErrInvalidValue
	}
	return box(n), nil
}

func parseFloat(raw string, bitSize int, box func(float64) any) (any, error) {
	n, err := strconv.ParseFloat(strings.TrimSpace(raw), bitSize)
	if err != nil {
		return nil, ErrInvalidValue
	}
	return box(n), nil
}

func parseBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, ErrInvalidValue
	}
}

func parseTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalidValue
}
