// Package tests provides map-based, deferred-execution HTTP test abstractions
// for Gin. Methods register expectations into maps/slots; nothing runs until
// .Run(). Both New (fluent) and NewFunc (declarative) build the same T type.
//
// Struct embedding all the way down — zero delegation methods:
//
//	Body       — all body state (request raw, dataType, response parsed)
//	HeaderMap  — request headers, with Header()/Set() methods
//	R          — request: embeds HeaderMap + Body
//	T          — test runner: embeds R
//
// t.JSON() IS Body.JSON() via embedding.
package fasttp

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// CheckFunc inspects the T after .Run() has executed the request.
// It uses require/assert on t.t for pass/fail signalling.
type CheckFunc func(tr *T)

// ---------- Body: all body state, reused via embedding ----------

// Body holds all body state for a request: the raw request payload, the
// data-type selector, and the parsed response body after .Run(). It is
// embedded in R (and therefore in T) so its methods are promoted for free.
type Body struct {
	raw      interface{} // request payload
	dataType string      // "json" | "xml" | "buffer" | "yaml" | ""
	parsed   interface{} // response body after Run() parses it
}

// Data returns the request payload marshaled to bytes according to dataType.
// []byte and string raw values pass through as-is; other values are marshaled
// per Type() (default: JSON).
func (b *Body) Data() []byte {
	if b.raw == nil {
		return nil
	}
	switch data := b.raw.(type) {
	case []byte:
		return data
	case string:
		return []byte(data)
	}
	switch b.dataType {
	case "xml":
		out, _ := xml.Marshal(b.raw)
		return out
	default: // "json", "buffer", "yaml", ""
		out, _ := json.Marshal(b.raw)
		return out
	}
}

// Type returns the dataType: "json" | "xml" | "buffer" | "yaml" | "".
func (b *Body) Type() string { return b.dataType }

// Parsed returns the response body parsed by Run() according to Type().
func (b *Body) Parsed() interface{} { return b.parsed }

// JSON sets dataType to "json" (last data-type call wins).
func (b *Body) JSON() *Body {
	b.dataType = "json"
	return b
}

// XML sets dataType to "xml" (last data-type call wins).
func (b *Body) XML() *Body {
	b.dataType = "xml"
	return b
}

// Buffer sets dataType to "buffer" (last data-type call wins).
func (b *Body) Buffer() *Body {
	b.dataType = "buffer"
	return b
}

// YAML sets dataType to "yaml" (stub — requires yaml dependency) (last call wins).
func (b *Body) YAML() *Body {
	b.dataType = "yaml"
	return b
}

// ---------- HeaderMap: map with methods ----------

// HeaderMap is a map of request headers to send.
type HeaderMap map[string]string

// Header sets header pairs onto the map: odd arguments are keys, even are
// values. Overwrites existing keys. Chainable.
func (h HeaderMap) Header(pairs ...string) HeaderMap {
	for i := 0; i+1 < len(pairs); i += 2 {
		h[pairs[i]] = pairs[i+1]
	}
	return h
}

// Set replaces all entries in the map with those from m. Chainable.
func (h HeaderMap) Set(m HeaderMap) HeaderMap {
	for k := range h {
		delete(h, k)
	}
	for k, v := range m {
		h[k] = v
	}
	return h
}

// ---------- R: request (embeds HeaderMap + Body) ----------

// R holds request configuration. It embeds HeaderMap (free Header(), Set())
// and Body (free JSON(), XML(), Buffer(), YAML(), Data(), Type(), Parsed()).
// It declares zero methods of its own — everything is promoted.
type R struct {
	Method string
	Path   string
	HeaderMap
	Body
}

// ---------- T: test runner (embeds R) ----------

// T holds all state for a single HTTP test. Request config is built via chain
// methods (maps/slots); nothing executes until Run().
//
// Embedding R gives free access to: Method, Path, HeaderMap, Body, Header(),
// Set(), JSON(), XML(), Buffer(), YAML(), Type(), Parsed().
// Note: T declares its own Data(fn) which shadows Body.Data() (the raw-bytes
// accessor stays reachable as t.Body.Data()).
type T struct {
	t *testing.T

	Name    string
	Router  *gin.Engine
	R       // Method, Path, HeaderMap, Body + promoted methods

	// Response state (set during/after Run)
	Recorder *httptest.ResponseRecorder

	// Check slots — executed in fixed order by Run(): status → data → headers.
	// Format validation happens during Run's parse phase per Body.Type().
	statusCheck  CheckFunc   // set by Status (last call wins)
	dataChecks   []CheckFunc // set by Data (appended, all run)
	headerChecks []CheckFunc // set by HeaderExists/HeaderEquals (appended, all run)
}

// defaultRouter is the Gin engine set by Start(). All New calls use it.
var defaultRouter *gin.Engine

// Start bootstraps the Gin router used by all tests. Call once in TestMain or
// at the top of each test file. Returns the router for handler registration.
func Start(r *gin.Engine) *gin.Engine {
	defaultRouter = r
	return r
}

// New creates a T in fluent/builder style. Nothing executes until .Run() is
// called. Every method returns its receiver for chaining.
func New(t *testing.T, name, path, method string, body []byte) *T {
	return &T{
		t:      t,
		Name:   name,
		Router: defaultRouter,
		R: R{
			Method:    method,
			Path:      path,
			HeaderMap: make(HeaderMap),
			Body:      Body{raw: body},
		},
		dataChecks:   make([]CheckFunc, 0),
		headerChecks: make([]CheckFunc, 0),
	}
}

// Status sets the expected response status code check. Last call wins.
func (t *T) Status(expected int) *T {
	t.statusCheck = checkStatus(expected)
	return t
}

// Data appends a callback that receives the parsed response and returns the
// expected data. The framework auto-asserts expected == actual after the
// callback returns — no manual assert.Equal needed.
//
// For maps/slices: receive a shallow copy, modify in-place, return it.
// For primitives/bytes: receive the actual, return what you expect.
//
//	tr.Data(func(data interface{}) interface{} {
//	    m := data.(map[string]interface{})
//	    m["name"] = "alice"  // set expected
//	    return m
//	})
func (t *T) Data(fn func(data interface{}) interface{}) *T {
	t.dataChecks = append(t.dataChecks, checkData(fn))
	return t
}

// HeaderExists appends a check that asserts a response header key is present
// and non-empty.
func (t *T) HeaderExists(key string) *T {
	t.headerChecks = append(t.headerChecks, checkResponseHeaderExists(key))
	return t
}

// HeaderEquals appends a check that asserts a response header has an exact value.
func (t *T) HeaderEquals(key, expected string) *T {
	t.headerChecks = append(t.headerChecks, checkResponseHeader(key, expected))
	return t
}

// Checks returns the ordered slice of all registered checks. Useful for
// equivalence testing (verifying New and NewFunc produce same checks).
func (t *T) Checks() []CheckFunc {
	var all []CheckFunc
	if t.statusCheck != nil {
		all = append(all, t.statusCheck)
	}
	all = append(all, t.dataChecks...)
	all = append(all, t.headerChecks...)
	return all
}

// Run marshals the request body via Body.Data(), executes the request, parses
// the response body according to Type() into Body.parsed (with per-type format
// validation), then runs all checks in fixed order:
//
//	statusCheck → dataChecks → headerChecks
//
// Returns the T for post-run inspection.
func (t *T) Run() *T {
	// Build and execute request
	bodyReader := strings.NewReader(string(t.Body.Data()))
	req := httptest.NewRequest(t.Method, t.Path, bodyReader)

	for k, v := range t.HeaderMap {
		req.Header.Set(k, v)
	}

	t.Recorder = httptest.NewRecorder()
	t.Router.ServeHTTP(t.Recorder, req)

	// Parse phase: format validation + decode response body into Body.parsed
	respBytes := t.Recorder.Body.Bytes()
	switch t.Type() {
	case "json":
		checkValidJSON()(t)
		var parsed interface{}
		_ = json.Unmarshal(respBytes, &parsed)
		t.parsed = parsed
	case "xml":
		checkValidXML()(t)
		var parsed interface{}
		_ = xml.Unmarshal(respBytes, &parsed)
		t.parsed = parsed
	case "buffer":
		checkValidBuffer()(t)
		t.parsed = respBytes
	case "yaml":
		checkValidYAML()(t)
		t.parsed = respBytes
	default:
		t.parsed = respBytes
	}

	// Check phase: fixed order execution
	if t.statusCheck != nil {
		t.statusCheck(t)
	}
	for _, check := range t.dataChecks {
		check(t)
	}
	for _, check := range t.headerChecks {
		check(t)
	}

	return t
}

// TestInfo holds the declarative test configuration for NewFunc.
type TestInfo struct {
	Method  string
	Status  int
	JSON    bool
	YAML    bool
	XML     bool
	Buffer  bool
	Headers HeaderMap
	Data    func(data interface{}) interface{}
}

// NewFunc creates a test using the declarative/single-call pattern.
// It internally builds a New chain from TestInfo, applies all settings via
// the same methods, and calls .Run(). Zero code duplication — every NewFunc
// is a New chain underneath.
func NewFunc(t *testing.T, name, path string, body []byte, fn func(*T) TestInfo) *T {
	tr := New(t, name, path, http.MethodGet, body)
	info := fn(tr)

	if info.Method != "" {
		tr.Method = info.Method
	}
	if info.Headers != nil {
		tr.HeaderMap.Set(info.Headers)
	}

	// Data-type: last bool wins, checked in priority order
	switch {
	case info.JSON:
		tr.JSON()
	case info.YAML:
		tr.YAML()
	case info.XML:
		tr.XML()
	case info.Buffer:
		tr.Buffer()
	}

	if info.Status != 0 {
		tr.Status(info.Status)
	}
	if info.Data != nil {
		tr.Data(info.Data)
	}

	return tr.Run()
}
