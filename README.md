<p align="center">
  <img src="https://skillicons.dev/icons?i=go" />
</p>
<br>

# gofilt

Lightweight, type-safe filter builder for Go. Parse filter parameters from struct tags or URL query params into a portable `Filter` AST, then apply it to GORM, `database/sql`, or any custom backend via adapter pattern.

## Features

- Parse filters from Go structs via reflection (`FromStruct`)
- Parse filters from `url.Values` / HTTP query string (`FromURL`)
- Whitelist allowed fields with `WithAllowedFields`
- Custom struct tag name with `WithTagName`
- 10 operators: `=`, `!=`, `>`, `>=`, `<`, `<=`, `LIKE`, `ILIKE`, `IN`, `BETWEEN`
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
    //   {Field: "name", Operator: "ILIKE", Value: "Budi"},
    //   {Field: "age",  Operator: ">=",    Value: 20},
    //   {Field: "status", Operator: "=",    Value: "active"},
    // ]
}
```

### From URL Query Params

```go
package main

import (
    "net/http"
    "github.com/noormaulida/gofilt"
)

func handler(w http.ResponseWriter, r *http.Request) {
    filter, _ := gofilt.FromURL(r.URL.Query(),
        gofilt.WithAllowedFields("name", "age", "status"),
    )
    // Query string: ?name[ilike]=Budi&age[gte]=20&status=active&limit=10&offset=20
    _ = filter.Limit  // 10
    _ = filter.Offset // 20
}
```

Supported URL syntax:

| Syntax | Operator |
|--------|----------|
| `name=Budi` | `=` (default) |
| `name[eq]=Budi` | `=` |
| `name[ne]=Budi` | `!=` |
| `age[gt]=20` | `>` |
| `age[gte]=21` | `>=` |
| `age[lt]=100` | `<` |
| `age[lte]=99` | `<=` |
| `name[like]=bud` | `LIKE` |
| `name[ilike]=bud` | `ILIKE` |
| `id[in]=1,2,3` | `IN` (auto-split comma) |
| `status=active,pending` | `IN` (auto-detect comma) |
| `limit=10` | Pagination |
| `offset=20` or `page=20` | Pagination |

### Apply with GORM

```go
import gofiltgorm "github.com/noormaulida/gofilt/adapter/gorm"

var users []User
db := gofiltgorm.Apply(tx, filter)
db.Find(&users)
```

### Build SQL for `database/sql`

```go
import gofiltsql "github.com/noormaulida/gofilt/adapter/sql"

where, args := gofiltsql.BuildWHERE(filter)
// where = "WHERE name ILIKE $1 AND age >= $2"
// args  = ["%Budi%", 20]
rows, _ := db.QueryContext(ctx, "SELECT * FROM users "+where, args...)
```

## Options

| Option | Description |
|--------|-------------|
| `WithTagName(name string)` | Override default struct tag (`"filt"`) |
| `WithAllowUnknown(bool)` | Reserved for future URL-parser strict mode |
| `WithAllowedFields(...string)` | Whitelist-only mode — drop any field not in the list |

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
├── gofilt.go              # Core types: Filter, Condition, Operator, Option
├── operator.go            # Operator lookup & aliases (public)
├── options.go             # Option constructors + defaultOptions
├── parser_struct.go       # FromStruct — reflect-based struct parser
├── parser_url.go          # FromURL — url.Values parser
├── adapter/
│   ├── gorm/              # gofiltgorm.Apply(*gorm.DB, *Filter)
│   └── sql/               # gofiltsql.BuildWHERE(*Filter) (Postgres $N style)
├── go.mod                 # Single module — one `go get` covers all packages
└── *_test.go              # Unit tests
```

Note: All packages (core + adapters) are part of one Go module. Adapter import paths stay `github.com/noormaulida/gofilt/adapter/gorm` and `.../adapter/sql` — publishing requires only one release tag `vX.Y.Z` for everything at once.

## Running Tests

```bash
go test -v ./...
```

## License

[MIT License](LICENSE)
