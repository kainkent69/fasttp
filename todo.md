# FastTP — TODO

> **Constraint:** `New` (fluent) and `NewFunc` (declarative) MUST be equivalent.
> Same capabilities, same underlying execution, same assertions. **`NewFunc` must use `New` internally — zero code dup.**

---

## Phase 4: API Redesign (promt2.md)

### 4.1 `HeaderMap` — type with methods
- [ ] `type HeaderMap map[string]string`
- [ ] `(h HeaderMap) Header(pairs ...string) HeaderMap` — odd=key, even=value, chainable
- [ ] `(h HeaderMap) Set(m HeaderMap) HeaderMap` — replace all entries

### 4.2 `Body` — request body struct
- [ ] `data interface{}` (private) — raw body data
- [ ] `dataType string` (private) — `"json"|"xml"|"buffer"|""`
- [ ] `(b *Body) Data() []byte` — marshal data to bytes based on dataType
- [ ] `(b *Body) Type() string` — return dataType
- [ ] `(b *Body) JSON() *Body` — set dataType="json", store raw data
- [ ] `(b *Body) XML() *Body` — set dataType="xml"
- [ ] `(b *Body) Buffer() *Body` — set dataType="buffer"

### 4.3 `R` — request (struct embedding)
- [ ] `Method string`
- [ ] `Path string`
- [ ] Embeds `HeaderMap` — gives `R.Header(pairs...)`, `R.Set(m)`
- [ ] `Body Body` — named field (NOT embedded, avoids `data` conflict with T.response data)
- [ ] `(r *R) JSON() *R` → `r.Body.JSON()`
- [ ] `(r *R) XML() *R` → `r.Body.XML()`
- [ ] `(r *R) Buffer() *R` → `r.Body.Buffer()`

### 4.4 `T` — test runner (final shape)
- [ ] `Req R` — named field (not embedded, avoids Body.data conflict)
- [ ] `Router *gin.Engine`
- [ ] `Recorder *httptest.ResponseRecorder`
- [ ] Private: `t *testing.T`, `data interface{}` (parsed response), `formatCheck/statusCheck/dataChecks/headerChecks`
- [ ] `(t *T) Header(pairs ...string) *T` → delegates to `t.Req.HeaderMap.Header()`
- [ ] `(t *T) SetHeader(pairs ...string) *T` → delegates to `t.Req.HeaderMap.Header()`
- [ ] `(t *T) Headers(m HeaderMap) *T` → delegates to `t.Req.HeaderMap.Set()`
- [ ] `(t *T) JSON() *T` → `t.Req.Body.JSON()` + register format check
- [ ] `(t *T) XML() *T` → `t.Req.Body.XML()` + register format check
- [ ] `(t *T) Buffer() *T` → `t.Req.Body.Buffer()` + register format check
- [ ] `(t *T) Status(code int) *T` — set status check slot
- [ ] `(t *T) Data(fn func(interface{})) *T` — append data callback
- [ ] `(t *T) HeaderExists(key string) *T` — response header check
- [ ] `(t *T) HeaderEquals(key, expected string) *T` — response header check
- [ ] `(t *T) Run() *T` — build request from Req, execute, parse, run checks
- [ ] `(t *T) Checks() []CheckFunc` — for equivalence testing

### 4.5 `New` + `NewFunc` (renamed + verified equivalent)
- [ ] `New(t, name, path, method string, body []byte) *T` (was NewTest)
- [ ] `NewFunc(t, name, path string, body []byte, fn func(*T) TestInfo) *T` (was NewFuncTest)
- [ ] `NewFunc` internally builds `New` chain — zero code dup
- [ ] Meta-test: `TestNewAndNewFuncAreEquivalent`

### 4.6 `status/` package
- [ ] `status/status.go` — typed constants: `Ok=200`, `Created=201`, `NoContent=204`, `BadRequest=400`, `Unauthorized=401`, `Forbidden=403`, `NotFound=404`, `Conflict=409`, `Unprocessable=422`, `TooMany=429`, `Internal=500`, `ServiceUnavailable=503`
- [ ] `status.Text(code int) string` — wraps `http.StatusText`

### 4.7 Update all consumers
- [ ] `tests/runner_test.go` — migrate to new API
- [ ] `sample/tests/*.go` — migrate all 93 tests to `New`/`NewFunc`/`status.*`
- [ ] `sample/app/` — verify no changes needed
- [ ] `tests/checks.go` — verify no changes needed
- [ ] All tests pass

---

## Phase 0: Project Setup (before any code)

## Phase 0: Project Setup (before any code)

- [x] Init Go module (`github.com/kainkent69/fasttp`, go 1.26.4)
- [x] Add dependencies: `gin-gonic/gin`, `stretchr/testify`
- [x] Create `CHANGELOG.md` — append-only, fine-grained per change: what, why, files touched, test impact
- [x] Create `./tests/` package directory

## Phase 1: Core Test Abstraction (Gin httptest)

> **Architecture:** Map-heavy, deferred execution. Methods register expectations into
> maps/slots. Nothing runs until `.Run()`. Like middleware/event pipeline:
> register handlers → fire request → iterate checks against response.
> **Implementation note:** Uses priority slots (formatCheck, statusCheck, dataChecks, headerChecks)
> instead of flat Checks slice to guarantee fixed execution order regardless of call order.

### 1.1 Core engine: `TestRunner` (single impl, both APIs share)
- [x] `HeaderMap map[string]string` — request headers, built by `.Header()` calls
- [x] `Body []byte` — request body
- [x] `Name, Path, Method string` — request identity
- [x] `Router *gin.Engine` — the Gin router under test
- [x] `Recorder *httptest.ResponseRecorder` — set after `.Run()`
- [x] `dataType string` — `""` (unset) | `"json"` | `"yaml"` | `"xml"` | `"buffer"`. Set by `.JSON()` etc.
- [x] `data interface{}` — **private** (unexported). Holds parsed response body after `.Run()`. Accessed via `.Data(fn)` callback.

### 1.2 Data-type methods (config + validation, no HTTP yet)
Each sets `dataType` AND sets the format-validation slot:

- [x] `.JSON() *TestRunner` → `t.dataType = "json"` + set `formatCheck` to `checkValidJSON`
- [x] `.YAML() *TestRunner` → `t.dataType = "yaml"` + set `formatCheck` to `checkValidYAML` (stub)
- [x] `.XML() *TestRunner` → `t.dataType = "xml"` + set `formatCheck` to `checkValidXML`
- [x] `.Buffer() *TestRunner` → `t.dataType = "buffer"` + set `formatCheck` to `checkValidBuffer`
- [x] Only one data-type can be set. Second call overwrites `dataType` + replaces format check.
- [x] If no data-type set, `.Data(fn)` still works — receives raw `[]byte` as `interface{}`

### 1.3 Map-based registration (deferred, no HTTP yet)
Remaining methods write to maps/slots, return `*TestRunner` for chaining:

- [x] `NewTest(t *testing.T, name, path, method string, data []byte) *TestRunner` — init empty maps + slots
- [x] `.Header(key, value string) *TestRunner` → `t.HeaderMap[key] = value`
- [x] `.Headers(h map[string]string) *TestRunner` → merge into `HeaderMap`
- [x] `.Status(expected int) *TestRunner` → set `statusCheck` slot (last wins)
- [x] `.Data(fn func(data interface{})) *TestRunner` → append to `dataChecks` (fn receives `t.data` after parse)
- [x] `.HeaderExists(key string) *TestRunner` → append to `headerChecks`
- [x] `.HeaderEquals(key, value string) *TestRunner` → append to `headerChecks`

### 1.4 `.Run()` — fire everything (the trigger)
- [x] Build Gin request from `Method`, `Path`, `Body`, `HeaderMap`
- [x] Execute via `httptest.NewRecorder` + `Router.ServeHTTP`
- [x] Store `Recorder`
- [x] **Parse phase:** if `dataType` is set, parse `Recorder.Body.Bytes()` into private `t.data`:
  - `"json"` → `json.Unmarshal` into `interface{}`
  - `"yaml"` → stub (dep not yet added)
  - `"xml"` → `xml.Unmarshal` into `interface{}`
  - `"buffer"` → raw `[]byte` as `interface{}`
- [x] **Check phase:** fixed order — formatCheck → statusCheck → dataChecks → headerChecks
- [x] If any check fails: `require.*` stops test immediately
- [x] Returns `*TestRunner` for post-run inspection

### 1.5 `NewFuncTest` — declarative sugar over `NewTest`
- [x] `TestInfo` struct: `Method string`, `Status int`, `JSON/YAML/XML/Buffer bool`, `Headers HeaderMap`, `Data func(interface{})`
- [x] `NewFuncTest(t *testing.T, name, path string, body []byte, fn func(*TestRunner) TestInfo)`
- [x] Internally: `tr := NewTest(...)` → read `TestInfo` → call same `.Header()`/`.Status()`/`.JSON()`/`.Data()` methods → `.Run()`
- [x] **Zero code duplication.** Every `NewFuncTest` call is a `NewTest` chain underneath.
- [x] **Must produce identical checks + same `dataType` + same `data` as equivalent `NewTest` chain.**

### 1.6 Check types (middleware/event handlers)
- [x] `type CheckFunc func(tr *TestRunner)` — signature (t accessed via tr.t)
- [x] `checkStatus(code int)` — `require.Equal(t, code, tr.Recorder.Code)`
- [x] `checkValidJSON()` — `json.Unmarshal` body into `json.RawMessage`, `require.NoError`
- [x] `checkValidYAML()` — stub (requires yaml dep)
- [x] `checkValidXML()` — `xml.NewDecoder` token loop, `require.NoError`
- [x] `checkValidBuffer()` — body is non-nil `[]byte`, `require.NotNil`
- [x] `checkData(fn func(interface{}))` — runs after parse phase, calls `fn(tr.data)`
- [x] `checkResponseHeader(key, value string)` — `require.Equal`
- [x] `checkResponseHeaderExists(key string)` — `require.NotEmpty`
- [x] Ordering: format → status → data callbacks → header checks (fixed, not call-order dependent)

### 1.7 Equivalence guarantee
- [x] `TestNewTestAndNewFuncTestAreEquivalent` — same inputs → same checks, same dataType, same outcome
- [x] `TestNewFuncTestInternallyUsesNewTest` — verifies NewFuncTest uses NewTest underneath

### 1.8 Internal assertions (testify)
- [x] Use `require.NoError` for hard-stop errors
- [x] Use `assert.Equal` for value comparisons (struct equality over field-by-field)
- [ ] Use `assert.JSONEq` for JSON string comparison (not yet needed)

### 1.9 Reusable helpers (`./tests/`)
- [x] `tests.Start()` — boot Gin router, stores in package-level `defaultRouter`
- [ ] `tests.Check(t *testing.T)` — reusable complex assertion blocks (Phase 3+)
- [ ] Group helpers by domain (auth, data, etc.)

---

## Phase 2: Agent Workflow System

### 2.1 CHANGELOG + fine-grained change tracking
- [ ] Every change appends to `CHANGELOG.md`: what changed, why, files, test impact, agent ID
- [ ] Format: `## [date] <change-summary>` with structured detail block
- [ ] `backtrack.md` — rollback log: record what was done so it can be undone
- [ ] Append-only: never rewrite history, always append

### 2.2 Orchestrator
- [ ] Main agent: parse tasks, dispatch to sub-agents
- [ ] Task queue with dependencies (sequential if shared state)
- [ ] Review findings: aggregate, check pass/fail, reprompt if needed
- [ ] Each review round → CHANGELOG entry with findings summary

### 2.3 Git worktree isolation
- [ ] Agent spawns in own worktree (`.claude/worktrees/agent-<id>`)
- [ ] Agent commits its work with detailed commit message
- [ ] Main agent checks out / reviews each worktree result

### 2.4 Tracking files (per agent)
- [ ] Append-only tracking file per worktree/branch (`tracking-<id>.md`)
- [ ] Each agent appends: what it did, decisions, test results, files changed
- [ ] `README[.branch].md` generated from tracking entries
- [ ] Tracking entries feed into root `CHANGELOG.md`

### 2.5 Execution modes
- [ ] Sequential: todo-list, no worktree needed, work in-place
- [ ] Parallel: independent agents in worktrees, run concurrently
- [ ] Hybrid: parallel for independent, sequential for dependent

### 2.6 Agent cycle
- [ ] Agent: checkout → implement → run tests → append tracking + CHANGELOG → notify main
- [ ] Main: review tracking → verify tests pass → merge or reprompt
- [ ] Reprompt includes fine-grained diff of what was tried vs what's needed

---

## Phase 3: Sample Project — HTTP Test Target ✅

### 3.1 Sample Gin app (`./sample/`)
- [x] Separated into `sample/app/` (importable package) + `sample/main.go` (thin wrapper)
- [x] In-memory store with mutex: CRUD items, delayed outputs, rate tracker, conflict tracker, failed status tracker
- [x] Endpoints: 19+ routes across 6 groups
  - [x] CRUD: GET/POST/PUT/PATCH/DELETE `/items`, `/items/:id`
  - [x] Content types: `/content/json`, `/content/xml`, `/content/text`, `/content/html`, `/content/bytes`
  - [x] Echo: `/echo`, `/headers`, `/reflect`
  - [x] Status codes: `/status/:code`
  - [x] Redirects: `/redirect` (301), `/redirect-temp` (302)
  - [x] Slow/delayed: `/slow` (500ms), `/stream` (async), `/stream/collect`
  - [x] Methods: `/methods` (GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS)
  - [x] Conflict: `/unique-items` (409 on duplicate name)
  - [x] Rate limited: `/rate-limited/data` (3 req/client, 429 exceeded)
  - [x] Auth: `/auth/profile`, `/auth/settings` (Bearer)
  - [x] Admin: `/admin/dashboard`, `/admin/failed-statuses` (admin scope)
  - [x] Edge: `/empty`, `/empty-body`, `/unicode`, `/large`, `/query-echo`, `/set-cookie`
- [x] Middleware: gin.Logger, gin.Recovery, CORS, auth (Bearer + scopes), admin-only, rate limiting

### 3.2 Sample tests (`./sample/tests/`) — 93 tests, all passing
- [x] `crud_test.go` (27): list empty/multi/filter, get found/not-found/bad-id/zero/negative, create success/bad-json/missing-name/empty/extra, put replace/not-found/bad-id, patch partial/not-found/tags, delete success/not-found/bad-id, lifecycle (5 sub-tests), unique-item/conflict, struct comparison
- [x] `content_test.go` (18): JSON nested, XML, text, HTML, bytes, echo object/empty/array, headers, reflect, unicode, large, query single/multi, empty JSON, no content
- [x] `auth_test.go` (13): no token, bad token, empty token, valid token, settings, admin user-token, admin admin-token, failed-statuses, rate-limit allowed (3x), rate-limit exceeded (429), different client, cookie
- [x] `method_test.go` (8): GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, CORS preflight
- [x] `edge_test.go` (27): all status codes (15 sub-tests), bad code, redirect 301/302, slow (≥500ms), stream delayed, stream multiple (3 batch), failed status tracking (404/400/401)
- [x] **Every single test uses `NewFuncTest` declarative style**
- [x] `assert.Equal` for struct comparisons, `assertMapEqual`/`assertMapHas` helpers
- [x] `t.Run()` sub-tests for groups (lifecycle, status codes)

### 3.3 fasttp abstractions demonstrated
- [x] `NewFuncTest()` declarative API — exclusive style in all 93 tests
- [x] `tests.Start()` injects Gin router via `app.SetupRouter()`
- [x] Content types: JSON, XML, Buffer (text/HTML/bytes)
- [x] Headers: custom request headers, header existence/value checks
- [x] Status codes: 200/201/204/301/302/400/401/403/404/405/409/422/429/500/503
- [x] No raw httptest — all tests use fasttp
