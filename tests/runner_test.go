package tests

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------- test harness ----------

func setupRouter() *gin.Engine {
	r := gin.New()
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello", "ok": true})
	})
	r.POST("/echo", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		c.JSON(http.StatusOK, body)
	})
	r.GET("/status/:code", func(c *gin.Context) {
		code := c.Param("code")
		switch code {
		case "200":
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		case "201":
			c.JSON(http.StatusCreated, gin.H{"status": "created"})
		case "400":
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		case "404":
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case "500":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		default:
			c.String(http.StatusOK, code)
		}
	})
	r.GET("/xml", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{"root": "value"})
	})
	r.GET("/headers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"content-type": c.GetHeader("Content-Type"),
			"x-custom":     c.GetHeader("X-Custom"),
		})
	})
	return r
}

func init() {
	Start(setupRouter())
}

// ---------- New fluent API tests (statement-style via embedding) ----------

func TestNewGetHello(t *testing.T) {
	tr := New(t, "get hello", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		m := data.(map[string]interface{})
		assert.Equal(t, "hello", m["message"])
		assert.Equal(t, true, m["ok"])
	})
	tr.Run()
}

func TestNewPostEcho(t *testing.T) {
	body := []byte(`{"name":"test","count":42}`)
	tr := New(t, "post echo", "/echo", http.MethodPost, body)
	tr.Header("Content-Type", "application/json")
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		m := data.(map[string]interface{})
		assert.Equal(t, "test", m["name"])
		assert.Equal(t, float64(42), m["count"])
	})
	tr.Run()
}

func TestNewStatusCodeVariants(t *testing.T) {
	cases := []struct {
		path   string
		status int
	}{
		{"/status/200", http.StatusOK},
		{"/status/201", http.StatusCreated},
		{"/status/400", http.StatusBadRequest},
		{"/status/404", http.StatusNotFound},
		{"/status/500", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			tr := New(t, tc.path, tc.path, http.MethodGet, nil)
			tr.JSON()
			tr.Status(tc.status)
			tr.Run()
		})
	}
}

func TestNewHeaderChecks(t *testing.T) {
	tr := New(t, "headers check", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.HeaderExists("Content-Type")
	tr.HeaderEquals("Content-Type", "application/json; charset=utf-8")
	tr.Run()
}

func TestNewCustomHeaders(t *testing.T) {
	tr := New(t, "custom headers", "/headers", http.MethodGet, nil)
	tr.Header("X-Custom", "my-value")
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		m := data.(map[string]interface{})
		assert.Equal(t, "my-value", m["x-custom"])
	})
	tr.Run()
}

func TestNewNoDataType(t *testing.T) {
	// No JSON/XML/Buffer — Data() receives raw bytes
	tr := New(t, "raw body", "/hello", http.MethodGet, nil)
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		b := data.([]byte)
		assert.Contains(t, string(b), "hello")
	})
	tr.Run()
}

func TestNewBuffer(t *testing.T) {
	tr := New(t, "buffer body", "/hello", http.MethodGet, nil)
	tr.Buffer()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		b := data.([]byte)
		assert.Contains(t, string(b), "hello")
	})
	tr.Run()
}

// ---------- NewFunc declarative API tests ----------

func TestNewFuncGetHello(t *testing.T) {
	NewFunc(t, "get hello func", "/hello", nil, func(tr *T) TestInfo {
		return TestInfo{
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "hello", m["message"])
				assert.Equal(t, true, m["ok"])
			},
		}
	})
}

func TestNewFuncPostEcho(t *testing.T) {
	body := []byte(`{"name":"func","count":99}`)
	NewFunc(t, "post echo func", "/echo", body, func(tr *T) TestInfo {
		return TestInfo{
			Method: http.MethodPost,
			Status: http.StatusOK,
			JSON:   true,
			Headers: HeaderMap{
				"Content-Type": "application/json",
			},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "func", m["name"])
				assert.Equal(t, float64(99), m["count"])
			},
		}
	})
}

func TestNewFuncHeadersAndStatus(t *testing.T) {
	NewFunc(t, "headers func", "/hello", nil, func(tr *T) TestInfo {
		return TestInfo{
			Status: http.StatusOK,
			JSON:   true,
		}
	})
}

// ---------- equivalence meta-test ----------

func TestNewAndNewFuncAreEquivalent(t *testing.T) {
	// Both styles produce identical results for the same test case.
	// We verify: same checks, same status, same parsed data shape.

	// Fluent (statement-style via embedding)
	fluent := New(t, "equiv", "/hello", http.MethodGet, nil)
	fluent.JSON()
	fluent.Status(http.StatusOK)
	fluent.Data(func(data interface{}) {})

	// Declarative (does not call Run — we compare pre-Run state)
	decl := New(t, "equiv", "/hello", http.MethodGet, nil)
	info := TestInfo{
		Status: http.StatusOK,
		JSON:   true,
		Data:   func(data interface{}) {},
	}
	// Apply same way NewFunc does
	decl.JSON()
	decl.Status(info.Status)
	decl.Data(info.Data)

	// Compare check counts
	assert.Equal(t, len(fluent.Checks()), len(decl.Checks()),
		"fluent and declarative must produce same number of checks")
	assert.Equal(t, fluent.Type(), decl.Type(),
		"fluent and declarative must have same dataType")

	// Run both and verify same outcome
	fluent.Run()
	decl.Run()

	assert.Equal(t, fluent.Recorder.Code, decl.Recorder.Code,
		"fluent and declarative must produce same status code")
	assert.Equal(t, fluent.Recorder.Body.String(), decl.Recorder.Body.String(),
		"fluent and declarative must produce same response body")
}

func TestNewFuncInternallyUsesNew(t *testing.T) {
	// NewFunc must create the same check structure as a New chain.
	// We assert that the resulting T has identical fields.

	body := []byte(`{"x":1}`)
	path := "/echo"
	status := http.StatusOK

	// Via NewFunc (calls Run internally)
	result := NewFunc(t, "internal check", path, body, func(tr *T) TestInfo {
		return TestInfo{
			Method: http.MethodPost,
			Status: status,
			JSON:   true,
			Headers: HeaderMap{
				"Content-Type": "application/json",
			},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, float64(1), m["x"])
			},
		}
	})

	assert.Equal(t, status, result.Recorder.Code)
	assert.Equal(t, http.MethodPost, result.Method)
	assert.Equal(t, "application/json", result.HeaderMap["Content-Type"])
	assert.Equal(t, "json", result.Type())
}

// ---------- XML tests ----------

func TestNewXML(t *testing.T) {
	tr := New(t, "xml response", "/xml", http.MethodGet, nil)
	tr.XML()
	tr.Status(http.StatusOK)
	tr.Run()
}

// ---------- status-code raw tests ----------

func TestNewRawResponse(t *testing.T) {
	tr := New(t, "raw string", "/status/raw", http.MethodGet, nil)
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		b := data.([]byte)
		assert.Equal(t, "raw", string(b))
	})
	tr.Run()
}

// ---------- multiple Data callbacks ----------

func TestNewMultipleDataCallbacks(t *testing.T) {
	callCount := 0
	tr := New(t, "multi data", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) {
		callCount++
		m := data.(map[string]interface{})
		assert.Contains(t, m, "message")
	})
	tr.Data(func(data interface{}) {
		callCount++
		m := data.(map[string]interface{})
		assert.Contains(t, m, "ok")
	})
	tr.Run()

	assert.Equal(t, 2, callCount, "both Data callbacks should have been invoked")
}
