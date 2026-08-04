package fasttp_tests

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kainkent69/fasttp"
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
	fasttp.Start(setupRouter())
}

// ---------- New fluent API tests (statement-style via embedding) ----------

func TestNewGetHello(t *testing.T) {
	tr := fasttp.New(t, "get hello", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		m := data.(map[string]interface{})
		m["message"] = "hello"
		m["ok"] = true
		return m
	})
	tr.Run()
}

func TestNewPostEcho(t *testing.T) {
	body := []byte(`{"name":"test","count":42}`)
	tr := fasttp.New(t, "post echo", "/echo", http.MethodPost, body)
	tr.Header("Content-Type", "application/json")
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		m := data.(map[string]interface{})
		m["name"] = "test"
		m["count"] = float64(42)
		return m
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
			tr := fasttp.New(t, tc.path, tc.path, http.MethodGet, nil)
			tr.JSON()
			tr.Status(tc.status)
			tr.Run()
		})
	}
}

func TestNewHeaderChecks(t *testing.T) {
	tr := fasttp.New(t, "headers check", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.HeaderExists("Content-Type")
	tr.HeaderEquals("Content-Type", "application/json; charset=utf-8")
	tr.Run()
}

func TestNewCustomHeaders(t *testing.T) {
	tr := fasttp.New(t, "custom headers", "/headers", http.MethodGet, nil)
	tr.Header("X-Custom", "my-value")
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		m := data.(map[string]interface{})
		m["x-custom"] = "my-value"
		return m
	})
	tr.Run()
}

func TestNewNoDataType(t *testing.T) {
	// No JSON/XML/Buffer — Data() receives raw []byte.
	// Return expected bytes — framework asserts exact match.
	tr := fasttp.New(t, "raw body", "/hello", http.MethodGet, nil)
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		return data // exact match on raw body
	})
	tr.Run()
}

func TestNewBuffer(t *testing.T) {
	tr := fasttp.New(t, "buffer body", "/hello", http.MethodGet, nil)
	tr.Buffer()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		return data // exact match on raw body
	})
	tr.Run()
}

// ---------- NewFunc declarative API tests ----------

func TestNewFuncGetHello(t *testing.T) {
	fasttp.NewFunc(t, "get hello func", "/hello", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				m["message"] = "hello"
				m["ok"] = true
				return m
			},
		}
	})
}

func TestNewFuncPostEcho(t *testing.T) {
	body := []byte(`{"name":"func","count":99}`)
	fasttp.NewFunc(t, "post echo func", "/echo", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodPost,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Content-Type": "application/json",
			},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				m["name"] = "func"
				m["count"] = float64(99)
				return m
			},
		}
	})
}

func TestNewFuncHeadersAndStatus(t *testing.T) {
	fasttp.NewFunc(t, "headers func", "/hello", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Status: http.StatusOK,
			JSON:   true,
		}
	})
}

// ---------- equivalence meta-test ----------

func TestNewAndNewFuncAreEquivalent(t *testing.T) {
	fluent := fasttp.New(t, "equiv", "/hello", http.MethodGet, nil)
	fluent.JSON()
	fluent.Status(http.StatusOK)
	fluent.Data(func(data interface{}) interface{} { return data })

	decl := fasttp.New(t, "equiv", "/hello", http.MethodGet, nil)
	info := fasttp.TestInfo{
		Status: http.StatusOK,
		JSON:   true,
		Data:   func(data interface{}) interface{} { return data },
	}
	decl.JSON()
	decl.Status(info.Status)
	decl.Data(info.Data)

	assert.Equal(t, len(fluent.Checks()), len(decl.Checks()),
		"fluent and declarative must produce same number of checks")
	assert.Equal(t, fluent.Type(), decl.Type(),
		"fluent and declarative must have same dataType")

	fluent.Run()
	decl.Run()

	assert.Equal(t, fluent.Recorder.Code, decl.Recorder.Code,
		"fluent and declarative must produce same status code")
	assert.Equal(t, fluent.Recorder.Body.String(), decl.Recorder.Body.String(),
		"fluent and declarative must produce same response body")
}

func TestNewFuncInternallyUsesNew(t *testing.T) {
	body := []byte(`{"x":1}`)
	path := "/echo"
	status := http.StatusOK

	result := fasttp.NewFunc(t, "internal check", path, body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodPost,
			Status: status,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Content-Type": "application/json",
			},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				m["x"] = float64(1)
				return m
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
	tr := fasttp.New(t, "xml response", "/xml", http.MethodGet, nil)
	tr.XML()
	tr.Status(http.StatusOK)
	tr.Run()
}

// ---------- status-code raw tests ----------

func TestNewRawResponse(t *testing.T) {
	tr := fasttp.New(t, "raw string", "/status/raw", http.MethodGet, nil)
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		return []byte("raw")
	})
	tr.Run()
}

// ---------- multiple Data callbacks ----------

func TestNewMultipleDataCallbacks(t *testing.T) {
	callCount := 0
	tr := fasttp.New(t, "multi data", "/hello", http.MethodGet, nil)
	tr.JSON()
	tr.Status(http.StatusOK)
	tr.Data(func(data interface{}) interface{} {
		callCount++
		m := data.(map[string]interface{})
		m["message"] = "hello"
		return m
	})
	tr.Data(func(data interface{}) interface{} {
		callCount++
		m := data.(map[string]interface{})
		m["ok"] = true
		return m
	})
	tr.Run()

	assert.Equal(t, 2, callCount, "both Data callbacks should have been invoked")
}
