# CHANGELOG

All notable changes to fasttp. Append-only, most recent first.

Format: `## [YYYY-MM-DD] <summary>` with detail block.

---

## [2026-08-04] Phase 1 core: TestRunner, NewTest, NewFuncTest, checks

### What
Implemented map-based deferred-execution HTTP test abstraction for Gin httptest.

### Files
- `tests/runner.go` — TestRunner struct, NewTest fluent API, NewFuncTest declarative API, Run()
- `tests/checks.go` — CheckFunc type, all check implementations (status, JSON, XML, buffer, headers, data)
- `tests/runner_test.go` — 16 tests covering both APIs, equivalence proof, edge cases

### Design
- All methods write to maps/slots. Nothing executes until `.Run()`.
- `.Run()` executes: build request → parse body per dataType → checks in fixed order (format → status → data → headers)
- `NewFuncTest` internally creates a `NewTest` chain — zero code duplication.
- Equivalence meta-test proves both APIs produce identical results.
- `*testing.T` stored in TestRunner, testify `require` for hard-stop assertions.

### Test impact
- 16/16 pass: fluent API, declarative API, status variants, headers, JSON/XML/buffer parsing, raw bytes, multiple Data callbacks, equivalence proof.

---

## [2026-08-04] Phase 3: Sample Gin app + integration tests using fasttp

### What
Created a complete sample Gin HTTP server with diverse endpoints and comprehensive integration tests using fasttp's NewTest and NewFuncTest APIs.

### Files
- `sample/main.go` — Gin app with CRUD /items, /headers, /echo, /status/:code, /slow, /admin/* (auth middleware)
- `sample/main_test.go` — 22 tests demonstrating fasttp: fluent API, declarative API, sub-tests with t.Run(), struct comparison, header checks, error cases

### Design
- In-memory store with mutex for concurrent-safe CRUD
- Auth middleware on /admin/* (Bearer token check)
- `resetStore()` helper for test isolation
- Tests use both `NewTest().Status().JSON().Data().Run()` (fluent) and `NewFuncTest()` (declarative) patterns
- Demonstrates `assert.Equal` struct comparison via `assertItemEqual` helper
- Lifecycle test uses `t.Run()` sub-tests: create → read → update → delete → verify

### Test impact
- 22/22 pass: CRUD, status codes, headers, echo, auth middleware, content-type, lifecycle

---

## [2026-08-04] Phase 3 expanded: Comprehensive sample app + 93 tests

### What
Rewrote sample project as a full HTTP test target demonstrating fasttp viability.
Every test uses `NewFuncTest` declarative style. Covers every HTTP edge case.

### Files
- `sample/app/store.go` — In-memory store: CRUD items, delayed output collector, rate tracker, conflict tracker, failed status tracker
- `sample/app/middleware.go` — CORS, auth (Bearer, scopes), admin-only, rate limiting
- `sample/app/handlers.go` — SetupRouter + all handlers: CRUD (GET/POST/PUT/PATCH/DELETE), content types, echo, reflect, status codes, redirects, slow, streaming, methods, conflict, auth routes, edge cases
- `sample/main.go` — Thin wrapper calling `app.SetupRouter().Run()`
- `sample/tests/helper_test.go` — Shared helpers, assertion utilities, `resetStore()`
- `sample/tests/crud_test.go` — 27 CRUD tests: list, filter, get, create, put, patch, delete, lifecycle, unique/conflict, struct comparison
- `sample/tests/content_test.go` — 18 content tests: JSON, XML, text, HTML, bytes, echo, headers, reflect, unicode, large response, query echo, empty/no-content
- `sample/tests/auth_test.go` — 13 auth tests: no token, bad token, valid token, admin scope, rate limiting (under/over/different-client), failed status tracking, cookie
- `sample/tests/method_test.go` — 8 method tests: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, CORS preflight
- `sample/tests/edge_test.go` — 27 edge tests: all status codes (15), redirects, slow, streaming/delayed output (multiple), failed status tracking

### Test impact
- 93/93 pass (16 tests + 77 sub-tests)
- Every test NewFuncTest style
- Covers: all HTTP methods, content types (JSON/XML/text/HTML/bytes), auth with scopes, rate limiting, streaming/delayed output, redirects, unicode, large responses, query params, cookies, failed status tracking

---

## [2026-08-04] Phase 4: API Redesign — Body, HeaderMap, R, T, New, NewFunc, status/

### What
Complete API redesign per promt2.md. Struct embedding all the way down — zero delegation methods.
`Body` reusable via embedding, `HeaderMap` with variadic methods, `R` request, `T` test runner.

### Files changed
- `tests/runner.go` — Full rewrite: `Body` (raw/dataType/parsed + Data/Type/Parsed/JSON/XML/Buffer/YAML), `HeaderMap` (Header(pairs...)/Set), `R` (embeds HeaderMap+Body), `T` (embeds R + check slots), `New`/`NewFunc`, `Start`, `TestInfo`
- `tests/checks.go` — All checks use `*T`, `checkData` reads via `Parsed()`, format checks auto-run in Run's parse phase
- `tests/runner_test.go` — 16 tests migrated: `NewTest`→`New`, `NewFuncTest`→`NewFunc`, chain style→statement style (embedding means `JSON()` returns `*Body`), `tr.dataType`→`tr.Type()`
- `sample/tests/*.go` — 93 tests migrated: `NewFuncTest`→`NewFunc`, `*fasttp.TestRunner`→`*fasttp.T`
- `status/status.go` — New package: 15 constants (Ok, Created, NoContent, Moved, Found, BadRequest, Unauthorized, Forbidden, NotFound, MethodNotAllowed, Conflict, Unprocessable, TooMany, Internal, ServiceUnavailable) + `Text()`

### Design highlights
- **Zero delegation.** `t.JSON()` IS `Body.JSON()` via embedding chain `T→R→Body`
- **Body reusable.** `raw` (request), `dataType`, `parsed` (response) all in one struct
- **HeaderMap.Header(pairs...string)** — variadic pairs for ergonomic use
- **Format validation auto.** Run's parse phase validates format per `Type()` — no manual check registration
- **NewFunc uses New internally** — single implementation, proven by equivalence test

### Test impact
- 109/109 pass (16 core + 93 sample)

