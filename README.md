<p align="center">
  <img src="https://skillicons.dev/icons?i=go" alt="Go" />
</p>

<div align="center">
  <h1 class="h1">gofilt</h1>
</div>

<p align="center">
  Lightweight, type-safe filter builder for Go.
</p>

<p align="center">
  <a href="https://github.com/noormaulida/gofilt/actions/workflows/go.yml">
    <img src="https://github.com/noormaulida/gofilt/actions/workflows/go.yml/badge.svg" alt="Go Tests" />
  </a>
  <a href="https://codecov.io/gh/noormaulida/gofilt">
    <img src="https://codecov.io/gh/noormaulida/gofilt/graph/badge.svg" alt="codecov" />
  </a>
  <a href="https://pkg.go.dev/github.com/noormaulida/gofilt">
    <img src="https://pkg.go.dev/badge/github.com/noormaulida/gofilt.svg" alt="Go Reference" />
  </a>
  <a href="https://github.com/noormaulida/gofilt/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/noormaulida/gofilt" alt="License" />
  </a>
</p>

## Features

- Parse filters from Go structs via reflection (`FromStruct`)
- Parse filters from `url.Values` / HTTP query string (`FromURL`)
- Whitelist allowed fields with `WithAllowedFields`
- Multi-field sorting with `sort=name,-created_at`
- Whitelist sortable fields with `WithAllowedSorts`
- Configurable pagination limits with `WithDefaultLimit` and `WithMaxLimit`
- Deterministic pagination validation with `ErrInvalidLimit` and `ErrInvalidOffset`
- Pagination metadata helper with `filter.Pagination()`, `Page()`, `NextOffset()`, `PreviousOffset()`
- Custom struct tag name with `WithTagName`
- 12 operators: `=`, `!=`, `>`, `>=`, `<`, `<=`, `LIKE`, `ILIKE`, `IN`, `BETWEEN`, `IS NULL`, `IS NOT NULL`
- Zero dependencies for the core module
- Optional adapters: GORM and plain SQL (PostgreSQL-style `$N` placeholders)

## Minimum Go Version

Go **1.22** or higher.

## Installation

```bash
go get github.com/noormaulida/gofilt@latest
```

This single command installs the core package plus the GORM adapter and SQL adapter together. No separate installs required.

## Quick Start

### From Struct

```go
package main

import "github.com/noormaulida/gofilt"

type UserQuery struct {
    Name   string `filt:"col:name;op:ilike"`
    Age    int    `filt:"col:age;op:gte"`
    Status string `filt:"col:status"`
}

func main() {
    req := UserQuery{Name: "Noor", Age: 20, Status: "active"}
    filter, _ := gofilt.FromStruct(req,
        gofilt.WithAllowedFields("name", "age", "status"),
    )
    // filter.Conditions = [
    //   {Field: "name", Operator: "ILIKE", Value: "Noor"},
    //   {Field: "age",  Operator: ">=",    Value: 20},
    //   {Field: "status", Operator: "=",    Value: "active"},
    // ]
}
```

### From URL Query Params

```go
package main

import (
    "errors"
    "net/http"

    "github.com/noormaulida/gofilt"
)

func handler(w http.ResponseWriter, r *http.Request) {
    filter, err := gofilt.FromURL(r.URL.Query(),
        gofilt.WithAllowedFields("name", "age", "status"),
        gofilt.WithAllowedSorts("name", "created_at"),
        gofilt.WithDefaultLimit(20),
        gofilt.WithMaxLimit(100),
    )
    if errors.Is(err, gofilt.ErrInvalidLimit) {
        http.Error(w, "invalid limit parameter", http.StatusBadRequest)
        return
    }
    if errors.Is(err, gofilt.ErrInvalidOffset) {
        http.Error(w, "invalid offset parameter", http.StatusBadRequest)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Query string:
    // ?name[ilike]=Noor&age[gte]=20&sort=name,-created_at&limit=10&offset=20
    _ = filter.Limit  // 10
    _ = filter.Offset // 20
    // filter.Sorts = [
    //   {Field: "name",       Direction: "ASC"},
    //   {Field: "created_at", Direction: "DESC"},
    // ]
}
```

Sorting is disabled until `WithAllowedSorts` is configured. Unknown sort
fields are ignored, preventing query parameters from becoming arbitrary SQL
identifiers.

Supported URL syntax:

| Syntax | Operator |
|--------|----------|
| `name=Noor` | `=` (default) |
| `name[eq]=Noor` | `=` |
| `name[ne]=Noor` | `!=` |
| `age[gt]=20` | `>` |
| `age[gte]=21` | `>=` |
| `age[lt]=100` | `<` |
| `age[lte]=99` | `<=` |
| `name[like]=bud` | `LIKE` |
| `name[ilike]=bud` | `ILIKE` |
| `id[in]=1,2,3` | `IN` (auto-split comma) |
| `status=active,pending` | `IN` (auto-detect comma) |
| `deleted_at[null]=true` | `IS NULL` (`isnull`, `is_null`, `is` are aliases; value is ignored) |
| `deleted_at[is]=null` | `IS NULL` |
| `deleted_at[notnull]=true` | `IS NOT NULL` (`isnotnull`, `not_null`, `is_not_null` are aliases) |
| `sort=name` | Sort by `name` ascending |
| `sort=name,-created_at` | Sort by `name` ascending, then `created_at` descending |
| `limit=10` | Pagination limit (defaults with `WithDefaultLimit`, capped by `WithMaxLimit`) |
| `offset=20` or `page=20` | Pagination offset |

Negative or non-integer values for `limit`, `offset`, or `page` return `ErrInvalidLimit` or `ErrInvalidOffset` deterministically.

### Typed URL values

Query values are strings until `WithFieldTypes` says otherwise. `FieldTypesFrom` reads column types from a query struct so the URL parser and the struct parser agree.

```go
type UserQuery struct {
    Age    int       `filt:"col:age;op:gte"`
    Active bool      `filt:"col:active"`
    Since  time.Time `filt:"col:created_at;op:gte"`
}

filter, err := gofilt.FromURL(r.URL.Query(),
    gofilt.WithFieldTypes(gofilt.FieldTypesFrom(UserQuery{})),
)
// ?age[gte]=20&active=true&created_at[gte]=2026-01-02T15:04:05Z
// Values are int(20), bool(true), and time.Time — not strings.
```

Supported types: `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint64`, `float32`, `float64`, `bool`, and `time.Time`.

`IN` lists are converted element by element (`id[in]=1,2,3` with an `int` field becomes `[]int`). Accepted time layouts are RFC3339 (with optional fractional seconds), `2006-01-02 15:04:05`, and `2006-01-02`. A value that does not match its field type returns `ErrInvalidValue`. Fields without a registered type stay strings. `IS NULL` and `IS NOT NULL` values are not converted.

### Pagination & Metadata

`Filter` provides a lightweight `Pagination()` helper to calculate page numbers and offsets:

```go
pagination := filter.Pagination()

page := pagination.Page()           // Current 1-based page number (e.g. 3)
next := pagination.NextOffset()     // Offset for next page (e.g. 30)
prev := pagination.PreviousOffset() // Offset for previous page (e.g. 10, clamped to 0)
```

> **Note**: `Total` record count is intentionally excluded from `Filter` and `Pagination`. Calculating total records requires a `COUNT(*)` database query, which is the responsibility of the repository or database layer.

### Apply with GORM

```go
import gofiltgorm "github.com/noormaulida/gofilt/adapter/gorm"

var users []User
db := gofiltgorm.Apply(tx, filter)
db.Find(&users)
// Adds ORDER BY name ASC, created_at DESC when filter.Sorts is populated.
```

### Build SQL for `database/sql`

```go
import gofiltsql "github.com/noormaulida/gofilt/adapter/sql"

where, args := gofiltsql.BuildWHERE(filter)
orderBy := gofiltsql.BuildORDER(filter)
// where = "WHERE name ILIKE $1 AND age >= $2"
// args  = ["%Noor%", 20]
// orderBy = "ORDER BY name ASC, created_at DESC"
rows, _ := db.QueryContext(ctx,
    "SELECT * FROM users "+where+" "+orderBy,
    args...,
)
```

## Options

| Option | Description |
|--------|-------------|
| `WithTagName(name string)` | Override default struct tag (`"filt"`) |
| `WithAllowUnknown(bool)` | Reserved for future URL-parser strict mode |
| `WithAllowedFields(...string)` | Whitelist-only mode — drop any field not in the list |
| `WithAllowedSorts(...string)` | Allow sorting only by listed fields; unknown fields are ignored |
| `WithDefaultLimit(limit int)` | Fallback limit when no limit is provided in input |
| `WithMaxLimit(max int)` | Cap the maximum allowed limit to guard against large queries |
| `WithFieldTypes(map[string]reflect.Type)` | Parse URL values into the listed Go types; omitted fields stay strings |

## Struct Tag Format

Tag format: `col:<column>;op:<operator>` — order does not matter. `op` is optional and defaults to `=`. Use `-` to ignore a field.

```go
type Query struct {
    A string `filt:"col:full_name;op:ilike"` // explicit
    B string `filt:"col:status"`              // op defaults to =
    C string `filt:"-"`                       // skip
    D string `custom:"col:x;op:ne"`           // use WithTagName("custom")
}
```

## Project Layout

```
gofilt/
├── gofilt.go              # Core types: Filter, Condition, Operator, Option, Errors
├── pagination.go          # Pagination metadata helper: Page, NextOffset, PreviousOffset
├── operator.go            # Operator lookup & aliases (public)
├── options.go             # Option constructors + defaultOptions
├── parser_struct.go       # FromStruct — reflect-based struct parser
├── parser_url.go          # FromURL — url.Values parser
├── value.go               # FieldTypesFrom and URL value conversion
├── adapter/
│   ├── gorm/              # Apply filters, sorting, and pagination to GORM
│   └── sql/               # Build WHERE and ORDER BY clauses
├── go.mod                 # Single module — one `go get` covers all packages
└── *_test.go              # Unit tests
```

Note: Core and adapters are released together under the same version.

## Running Tests

```bash
go test -v ./...
```

## License

[MIT License](LICENSE)
