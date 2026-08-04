# fasttp

Map-based, deferred-execution HTTP test framework for http but mainly created for [Gin](https://github.com/gin-gonic/gin).
Write integration tests that read like intent — no boilerplate `httptest` wrangling.

**140 tests. Zero delegation. Struct embedding all the way down.**

```go
func TestGetUser(t *testing.T) {
    fasttp.NewFunc(t, "get user", "/users/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
        return fasttp.TestInfo{
            Status: 200,
            JSON:   true,
            Data: func(data interface{}) interface{} {
                m := data.(map[string]interface{})
                m["name"] = "alice"  // set what you expect
                return m              // framework asserts expected == actual
            },
        }
    })
}
```

## Why

Gin's `httptest` is verbose. Every test builds a request, a recorder, serves, reads bytes, unmarshals, asserts. fasttp collapses this into a declaration of what you expect:

- **No `httptest.NewRequest`** — set Method, Path, Body directly
- **No `json.Unmarshal`** — `JSON()` validates + parses automatically
- **No `w.Code` checks** — `Status(200)` asserts via testify `require`
- **No header boilerplate** — `Header("Key","Val")` sets request headers, `HeaderEquals("Key","Val")` asserts response headers

## Architecture

Struct embedding gives promoted methods for free — no delegation boilerplate:

```
Body       — raw request payload, dataType selector, parsed response
  ↓ embedded in
R          — request: Method, Path + HeaderMap + Body
  ↓ embedded in
T          — test runner: R + check slots + Run()
```

`t.JSON()` **is** `Body.JSON()` via embedding. No wrapper methods.

## Install

```bash
go get github.com/kainkent69/fasttp
```

Requires `gin-gonic/gin` and `stretchr/testify`.

## Quick start

### Fluent API (`New`)

Statement-style via embedding — each method call is a separate statement:

```go
import (
    "testing"
    fasttp "github.com/kainkent69/fasttp/tests"
)

func TestHello(t *testing.T) {
    // Boot once in TestMain or init()
    fasttp.Start(myRouter)

    tr := fasttp.New(t, "hello", "/hello", "GET", nil)
    tr.JSON()
    tr.Status(200)
    tr.Data(func(data interface{}) interface{} {
        m := data.(map[string]interface{})
        m["message"] = "world"  // set what you expect
        return m                 // framework asserts expected == actual
    })
    tr.Run()
}
```

### Declarative API (`NewFunc`)

Same result, single call — `NewFunc` builds a `New` chain internally:

```go
func TestHello(t *testing.T) {
    fasttp.NewFunc(t, "hello", "/hello", nil, func(tr *fasttp.T) fasttp.TestInfo {
        return fasttp.TestInfo{
            Status: 200,
            JSON:   true,
            Data: func(data interface{}) interface{} {
                m := data.(map[string]interface{})
                m["message"] = "world"  // set expected
                return m                 // auto-asserted
            },
        }
    })
}
```

**Return `nil`** to skip auto-assert and use manual testify assertions instead (for complex checks like `assert.Contains`, `assert.GreaterOrEqual`, loops).

`New` and `NewFunc` produce identical test results. An [equivalence meta-test](tests/runner_test.go) proves this.

## API reference

### Body

All body state lives here — request payload, format selector, parsed response:

| Method | Returns | Description |
|---|---|---|
| `JSON()` | `*Body` | Set request body format to JSON. Auto-validates response is valid JSON in `Run()`. |
| `XML()` | `*Body` | Set request body format to XML. |
| `Buffer()` | `*Body` | Raw bytes mode — no parsing. |
| `Data()` | `[]byte` | Marshal request payload to bytes per selected format. |
| `Type()` | `string` | Return current dataType (`"json"`, `"xml"`, `"buffer"`, `""`). |
| `Parsed()` | `interface{}` | Return parsed response body (set by `Run()`). |

Body is embedded in both `R` and `T`. Call `.JSON()` directly on your test.

### HeaderMap

Request headers as `map[string]string` with ergonomic methods:

```go
// Variadic pairs — odd=key, even=value
t.Header("Content-Type", "application/json", "Authorization", "Bearer xyz")

// Replace all headers
t.Set(fasttp.HeaderMap{"Accept": "application/json"})
```

| Method | Returns | Description |
|---|---|---|
| `Header(pairs ...string)` | `HeaderMap` | Set headers by key-value pairs. Overwrites existing keys. |
| `Set(m HeaderMap)` | `HeaderMap` | Replace all entries with `m`. |

### R (Request)

Embeds `HeaderMap` and `Body`. Zero methods of its own — everything is promoted.

```go
type R struct {
    Method string
    Path   string
    HeaderMap       // → Header(), Set()
    Body            // → JSON(), XML(), Buffer(), Data(), Type(), Parsed()
}
```

### T (Test runner)

Embeds `R`. Build assertions via methods, fire via `Run()`.

**Exported fields:**

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Test name (set by `New`). |
| `Router` | `*gin.Engine` | The Gin router under test. |
| `R` | `R` | Embedded request — gives promoted access to `Method`, `Path`, `HeaderMap`, `Body` + all their methods. |
| `Recorder` | `*httptest.ResponseRecorder` | Response recorder, set after `Run()`. |

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `Status(code int)` | `*T` | Assert response status code. Last call wins. |
| `Data(fn func(interface{}) interface{})` | `*T` | Callback receives parsed response, returns expected data. Framework auto-asserts `expected == actual`. Return `nil` to skip auto-assert (manual assertions). **Shadows `Body.Data()`** — raw bytes via `t.Body.Data()`. |
| `HeaderExists(key string)` | `*T` | Assert response header is present and non-empty. |
| `HeaderEquals(key, expected string)` | `*T` | Assert response header has exact value. |
| `Run()` | `*T` | Marshal body → execute request → parse response → run checks. |
| `Checks()` | `[]CheckFunc` | Return ordered check list. For equivalence testing. |

**Run() execution order:** format validation → Status → Data callbacks → Header checks.

### New / NewFunc

```go
func New(t *testing.T, name, path, method string, body []byte) *T
func NewFunc(t *testing.T, name, path string, body []byte, fn func(*T) TestInfo) *T
```

`NewFunc` creates a `New` chain internally — single implementation, zero duplication.

### TestInfo

```go
type TestInfo struct {
    Method  string              // HTTP method (overrides default GET)
    Status  int                 // Expected status code
    JSON    bool                // Parse response as JSON
    XML     bool                // Parse response as XML
    Buffer  bool                // Raw bytes mode
    YAML    bool                // Parse response as YAML (stub)
    Headers HeaderMap           // Request headers
    Data    func(interface{})   // Assertion callback on parsed response
}
```

### Start

```go
func Start(r *gin.Engine) *gin.Engine
```

Call once before tests — sets the Gin router all `New` calls use. Returns the router for handler registration.

### Status constants

Import `github.com/kainkent69/fasttp/status`:

```go
status.Ok               // 200
status.Created          // 201
status.Accepted         // 202
status.NoContent        // 204
status.Moved            // 301
status.Found            // 302
status.BadRequest       // 400
status.Unauthorized     // 401
status.Forbidden        // 403
status.NotFound         // 404
status.MethodNotAllowed // 405
status.Conflict         // 409
status.Unprocessable    // 422
status.TooMany          // 429
status.Internal         // 500
status.ServiceUnavailable // 503
```

All constants are typed `int`. Use `status.Text(code)` for human-readable names.

## Project structure

```
fasttp/
  tests/                 # Core framework
    runner.go            # Body, HeaderMap, R, T, New, NewFunc, Start, TestInfo
    checks.go            # CheckFunc + all check implementations
    runner_test.go       # 16 meta-tests including equivalence proof
  status/
    status.go            # HTTP status constants + Text()
  sample/                # Demo 1: E-commerce-ish CRUD
    app/                 # Router, handlers, middleware, store
    tests/               # 93 tests (CRUD, content types, auth, rate limiting, edge cases)
  sample2/               # Demo 2: chi REST example ported to Gin
    app/                 # Articles API with middleware (ArticleCtx, AdminOnly, paginate)
    tests/               # 31 tests (CRUD, search, slug lookup, admin, lifecycle)
  CHANGELOG.md           # Append-only, fine-grained change log
  todo.md                # Task tracking
```

## Running tests

```bash
# All 140 tests
go test ./...

# Just the framework
go test ./tests/ -v

# Sample 1 (93 tests)
go test ./sample/tests/ -v

# Sample 2 (31 tests)
go test ./sample2/tests/ -v
```

## Content type support

| Type | Method | Response validation | Parse |
|---|---|---|---|
| JSON | `.JSON()` | `json.Unmarshal` → valid JSON required | `map[string]interface{}` |
| XML | `.XML()` | Token decode loop → valid XML required | `interface{}` |
| Raw bytes | `.Buffer()` | Body not nil | `[]byte` |
| None | (default) | None | `[]byte` raw |


