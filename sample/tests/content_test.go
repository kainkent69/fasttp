package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kainkent69/fasttp"
)

// ---------- Content-Type JSON ----------

func TestContentTypeJSON(t *testing.T) {
	fasttp.NewFunc(t, "json content", "/content/json", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Accept": "application/json",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "type", "json")
				nested := assertMapHas(t, data, "nested")
				nestedMap := nested.(map[string]interface{})
				assert.Equal(t, true, nestedMap["deep"])
				return nil
			},
		}
	})
}

// ---------- Content-Type XML ----------

func TestContentTypeXML(t *testing.T) {
	fasttp.NewFunc(t, "xml content", "/content/xml", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			XML:    true,
		}
	})
}

// ---------- Content-Type text/plain ----------

func TestContentTypeText(t *testing.T) {
	fasttp.NewFunc(t, "text content", "/content/text", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			Buffer: true,
			Data: func(data interface{}) interface{} {
				b := data.([]byte)
				assert.Equal(t, "plain text response", string(b))
				return nil
			},
		}
	})
}

// ---------- Content-Type text/html ----------

func TestContentTypeHTML(t *testing.T) {
	fasttp.NewFunc(t, "html content", "/content/html", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			Buffer: true,
			Data: func(data interface{}) interface{} {
				b := data.([]byte)
				assert.Contains(t, string(b), "<html>")
				return nil
			},
		}
	})
}

// ---------- Content-Type octet-stream ----------

func TestContentTypeBytes(t *testing.T) {
	fasttp.NewFunc(t, "bytes content", "/content/bytes", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			Buffer: true,
			Data: func(data interface{}) interface{} {
				b := data.([]byte)
				assert.Equal(t, []byte{0x00, 0x01, 0x02, 0xFF}, b)
				return nil
			},
		}
	})
}

// ---------- Echo body ----------

func TestEchoJSONBody(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{
		"message": "hello",
		"count":   42,
		"nested":  map[string]interface{}{"deep": "value"},
	})
	fasttp.NewFunc(t, "echo json", "/echo", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "message", "hello")
				assertMapEqual(t, data, "count", float64(42))
				nested := assertMapHas(t, data, "nested")
				nestedMap := nested.(map[string]interface{})
				assert.Equal(t, "value", nestedMap["deep"])
				return nil
			},
		}
	})
}

func TestEchoEmptyObject(t *testing.T) {
	fasttp.NewFunc(t, "echo empty", "/echo", []byte(`{}`), func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				assert.Empty(t, m)
				return nil
			},
		}
	})
}

func TestEchoArrayBody(t *testing.T) {
	// Echo endpoint binds to map[string]interface{}, so arrays become
	// {"array": [1,2,3]} — wrap in object for proper echo.
	body := []byte(`{"values": [1,2,3]}`)
	fasttp.NewFunc(t, "echo array", "/echo", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				vals := m["values"].([]interface{})
				assert.Len(t, vals, 3)
				return nil
			},
		}
	})
}

// ---------- Echo headers ----------

func TestEchoHeaders(t *testing.T) {
	fasttp.NewFunc(t, "echo headers", "/headers", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"X-Custom": "my-value",
				"Accept":   "text/plain",
			},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				assert.Equal(t, "my-value", m["X-Custom"])
				return nil
			},
		}
	})
}

// ---------- Reflect full ----------

func TestReflectFull(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{"key": "val"})
	fasttp.NewFunc(t, "reflect", "/reflect", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				assert.Equal(t, "POST", m["method"])
				assert.Equal(t, "/reflect", m["path"])
				bodyVal := m["body"].(map[string]interface{})
				assert.Equal(t, "val", bodyVal["key"])
				return nil
			},
		}
	})
}

// ---------- Response header checks ----------

func TestContentTypeHeaderPresent(t *testing.T) {
	fasttp.NewFunc(t, "ct header", "/content/json", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
		}
	})
	// Note: HeaderExists verified inside Run via Check ordering
}

func TestJSONContentTypeHeaderValue(t *testing.T) {
	fasttp.NewFunc(t, "ct header value", "/items", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
		}
	})
}

// ---------- Unicode ----------

func TestUnicodeResponse(t *testing.T) {
	fasttp.NewFunc(t, "unicode", "/unicode", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				assert.Equal(t, "こんにちは世界", m["message"])
				assert.Equal(t, "🚀✨", m["emoji"])
				return nil
			},
		}
	})
}

// ---------- Large response ----------

func TestLargeResponse(t *testing.T) {
	fasttp.NewFunc(t, "large response", "/large", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				items := m["items"].([]interface{})
				assert.Len(t, items, 100)
				return nil
			},
		}
	})
}

// ---------- Query echo ----------

func TestQueryEchoSingle(t *testing.T) {
	fasttp.NewFunc(t, "query single", "/query-echo?q=search&page=1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "q", "search")
				assertMapEqual(t, data, "page", "1")
				return nil
			},
		}
	})
}

func TestQueryEchoMultipleSort(t *testing.T) {
	fasttp.NewFunc(t, "query multi sort", "/query-echo?sort=name&sort=price", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				sortVals := m["sort"].([]interface{})
				assert.Len(t, sortVals, 2)
				return nil
			},
		}
	})
}

// ---------- Empty / no-content ----------

func TestEmptyJSONResponse(t *testing.T) {
	fasttp.NewFunc(t, "empty json", "/empty", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				assert.Nil(t, data)
				return nil
			},
		}
	})
}

func TestNoContentResponse(t *testing.T) {
	// 204 No Content has empty body — don't use Buffer (body bytes are nil)
	fasttp.NewFunc(t, "no content", "/empty-body", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNoContent,
		}
	})
}
