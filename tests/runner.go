// Package tests provides map-based, deferred-execution HTTP test abstractions
// for Gin. Methods register expectations into maps/slots; nothing runs until
// .Run(). Both NewTest (fluent) and NewFuncTest (declarative) delegate to the
// same TestRunner engine.
package tests

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// HeaderMap is a map of request headers to send.
type HeaderMap map[string]string

// CheckFunc inspects the TestRunner after .Run() has executed the request.
// It uses require/assert on tr.t for pass/fail signalling.
type CheckFunc func(tr *TestRunner)

// TestRunner holds all state for a single HTTP test. Request config is built
// via chain methods (maps/slots). Nothing executes until Run().
type TestRunner struct {
	t *testing.T

	// Request config (set before Run)
	Name      string
	Path      string
	Method    string
	Body      []byte
	HeaderMap HeaderMap
	Router    *gin.Engine

	// Response state (set during/after Run)
	Recorder *httptest.ResponseRecorder

	// Check slots — executed in fixed order by Run()
	formatCheck  CheckFunc   // set by JSON/YAML/XML/Buffer (last call wins)
	statusCheck  CheckFunc   // set by Status (last call wins)
	dataChecks   []CheckFunc // set by Data (appended, all run)
	headerChecks []CheckFunc // set by HeaderExists/HeaderEquals (appended, all run)

	// Data handling
	dataType string      // "json" | "xml" | "buffer" | ""
	data     interface{} // parsed response body (private), set during Run parse phase
}

// defaultRouter is the Gin engine set by Start(). All NewTest calls use it.
var defaultRouter *gin.Engine

// Start bootstraps the Gin router used by all tests. Call once in TestMain or
// at the top of each test file. Returns the router for handler registration.
func Start(r *gin.Engine) *gin.Engine {
	defaultRouter = r
	return r
}

// NewTest creates a TestRunner in fluent/builder style. Nothing executes until
// .Run() is called. Every method returns the same runner for chaining.
func NewTest(t *testing.T, name, path, method string, body []byte) *TestRunner {
	return &TestRunner{
		t:           t,
		Name:        name,
		Path:        path,
		Method:      method,
		Body:        body,
		HeaderMap:   make(HeaderMap),
		Router:      defaultRouter,
		dataChecks:  make([]CheckFunc, 0),
		headerChecks: make([]CheckFunc, 0),
	}
}

// Header sets a single request header into HeaderMap. Overwrites existing key.
func (tr *TestRunner) Header(key, value string) *TestRunner {
	tr.HeaderMap[key] = value
	return tr
}

// Headers merges a map of request headers into HeaderMap. Overwrites existing keys.
func (tr *TestRunner) Headers(h HeaderMap) *TestRunner {
	for k, v := range h {
		tr.HeaderMap[k] = v
	}
	return tr
}

// Status sets the expected response status code check. Last call wins.
func (tr *TestRunner) Status(expected int) *TestRunner {
	tr.statusCheck = checkStatus(expected)
	return tr
}

// JSON sets dataType to "json" and registers a format-validation check.
// During Run(), the body is parsed as JSON into the private .data field
// before any Data() callbacks execute. Last data-type call wins.
func (tr *TestRunner) JSON() *TestRunner {
	tr.dataType = "json"
	tr.formatCheck = checkValidJSON()
	return tr
}

// YAML sets dataType to "yaml" and registers a format-validation check.
// Last data-type call wins.
// TODO: requires gopkg.in/yaml.v3 dependency.
func (tr *TestRunner) YAML() *TestRunner {
	tr.dataType = "yaml"
	tr.formatCheck = checkValidYAML()
	return tr
}

// XML sets dataType to "xml" and registers a format-validation check.
// During Run(), the body is parsed as XML into the private .data field.
// Last data-type call wins.
func (tr *TestRunner) XML() *TestRunner {
	tr.dataType = "xml"
	tr.formatCheck = checkValidXML()
	return tr
}

// Buffer sets dataType to "buffer" and registers a format-validation check.
// The response body is stored as raw []byte in the private .data field.
// Last data-type call wins.
func (tr *TestRunner) Buffer() *TestRunner {
	tr.dataType = "buffer"
	tr.formatCheck = checkValidBuffer()
	return tr
}

// Data appends a callback that receives the already-parsed response data.
// Multiple Data() calls are all executed in registration order.
// If no dataType was set, fn receives the raw []byte body.
func (tr *TestRunner) Data(fn func(data interface{})) *TestRunner {
	tr.dataChecks = append(tr.dataChecks, checkData(fn))
	return tr
}

// HeaderExists appends a check that asserts a response header key is present
// and non-empty.
func (tr *TestRunner) HeaderExists(key string) *TestRunner {
	tr.headerChecks = append(tr.headerChecks, checkResponseHeaderExists(key))
	return tr
}

// HeaderEquals appends a check that asserts a response header has an exact value.
func (tr *TestRunner) HeaderEquals(key, expected string) *TestRunner {
	tr.headerChecks = append(tr.headerChecks, checkResponseHeader(key, expected))
	return tr
}

// Checks returns the ordered slice of all registered checks. Useful for
// equivalence testing (verifying NewTest and NewFuncTest produce same checks).
func (tr *TestRunner) Checks() []CheckFunc {
	var all []CheckFunc
	if tr.formatCheck != nil {
		all = append(all, tr.formatCheck)
	}
	if tr.statusCheck != nil {
		all = append(all, tr.statusCheck)
	}
	all = append(all, tr.dataChecks...)
	all = append(all, tr.headerChecks...)
	return all
}

// Run builds the Gin request from maps, executes it, parses the response body
// according to dataType, then runs all checks in fixed order:
//
//	formatCheck → statusCheck → dataChecks → headerChecks
//
// Returns the TestRunner for post-run inspection.
func (tr *TestRunner) Run() *TestRunner {
	// Build and execute request
	bodyReader := strings.NewReader(string(tr.Body))
	req := httptest.NewRequest(tr.Method, tr.Path, bodyReader)

	for k, v := range tr.HeaderMap {
		req.Header.Set(k, v)
	}

	tr.Recorder = httptest.NewRecorder()
	tr.Router.ServeHTTP(tr.Recorder, req)

	// Parse phase: decode response body into private .data
	respBytes := tr.Recorder.Body.Bytes()
	switch tr.dataType {
	case "json":
		var parsed interface{}
		_ = json.Unmarshal(respBytes, &parsed)
		tr.data = parsed
	case "xml":
		var parsed interface{}
		_ = xml.Unmarshal(respBytes, &parsed)
		tr.data = parsed
	case "buffer":
		tr.data = respBytes
	default:
		tr.data = respBytes
	}

	// Check phase: fixed order execution
	if tr.formatCheck != nil {
		tr.formatCheck(tr)
	}
	if tr.statusCheck != nil {
		tr.statusCheck(tr)
	}
	for _, check := range tr.dataChecks {
		check(tr)
	}
	for _, check := range tr.headerChecks {
		check(tr)
	}

	return tr
}

// TestInfo holds the declarative test configuration for NewFuncTest.
type TestInfo struct {
	Method  string
	Status  int
	JSON    bool
	YAML    bool
	XML     bool
	Buffer  bool
	Headers HeaderMap
	Data    func(data interface{})
}

// NewFuncTest creates a test using the declarative/single-call pattern.
// It internally builds a NewTest chain from TestInfo, applies all settings
// via the same methods, and calls .Run(). Zero code duplication — every
// NewFuncTest is a NewTest chain underneath.
func NewFuncTest(t *testing.T, name, path string, body []byte, fn func(*TestRunner) TestInfo) *TestRunner {
	tr := NewTest(t, name, path, http.MethodGet, body)
	info := fn(tr)

	if info.Method != "" {
		tr.Method = info.Method
	}
	if info.Headers != nil {
		tr.Headers(info.Headers)
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
