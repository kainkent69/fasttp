package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kainkent69/fasttp"
)

// ---------- Status codes ----------

func TestStatusCodes(t *testing.T) {
	cases := []struct {
		code int
		json bool // false for statuses that return no body
	}{
		{200, true}, {201, true},
		{204, false},               // No Content has empty body
		{301, false}, {302, false}, // Redirects have no body
		{400, true}, {401, true}, {403, true}, {404, true},
		{405, true}, {409, true}, {422, true},
		{429, true}, {500, true}, {503, true},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.code), func(t *testing.T) {
			info := fasttp.TestInfo{
				Method: http.MethodGet,
				Status: tc.code,
			}
			if tc.json {
				info.JSON = true
			}
			fasttp.NewFunc(t, "status "+http.StatusText(tc.code),
				"/status/"+itoa(tc.code), nil,
				func(tr *fasttp.T) fasttp.TestInfo {
					return info
				})
		})
	}
}

func TestStatusBadCode(t *testing.T) {
	fasttp.NewFunc(t, "bad status code", "/status/999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusBadRequest,
			JSON:   true,
		}
	})
}

// ---------- Redirect ----------

func TestRedirectPermanent(t *testing.T) {
	fasttp.NewFunc(t, "redirect 301", "/redirect", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusMovedPermanently,
		}
	})
}

func TestRedirectTemp(t *testing.T) {
	fasttp.NewFunc(t, "redirect 302", "/redirect-temp", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusFound,
		}
	})
}

// ---------- Slow response ----------

func TestSlowResponse(t *testing.T) {
	start := time.Now()
	fasttp.NewFunc(t, "slow", "/slow", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "ready", true)
				return nil
			},
		}
	})
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 500*time.Millisecond, "slow endpoint should take at least 500ms")
}

// ---------- Streaming / delayed output ----------

func TestStreamDelayedOutput(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{
		"data":     "async-message",
		"delay_ms": 50,
	})
	// Queue the delayed write
	fasttp.NewFunc(t, "stream post", "/stream", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusAccepted,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "queued", true)
				return nil
			},
		}
	})

	// Wait for goroutine to write
	time.Sleep(100 * time.Millisecond)

	// Collect
	fasttp.NewFunc(t, "stream collect", "/stream/collect", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				count := m["count"].(float64)
				assert.GreaterOrEqual(t, count, float64(1))
				entries := m["entries"].([]interface{})
				found := false
				for _, e := range entries {
					entry := e.(map[string]interface{})
					if entry["data"] == "async-message" {
						found = true
						break
					}
				}
				assert.True(t, found, "should find async-message in collected entries")
				return nil
			},
		}
	})
}

func TestStreamMultipleDelayed(t *testing.T) {
	for i := 0; i < 3; i++ {
		body := jsonMarshal(map[string]interface{}{
			"data":     "batch-" + itoa(i),
			"delay_ms": 10,
		})
		fasttp.NewFunc(t, "stream batch", "/stream", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPost,
				Status:  http.StatusAccepted,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			}
		})
	}
	time.Sleep(80 * time.Millisecond)

	fasttp.NewFunc(t, "stream collect all", "/stream/collect", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				entries := m["entries"].([]interface{})
				assert.GreaterOrEqual(t, len(entries), 3)
				return nil
			},
		}
	})
}

// ---------- Failed status tracking ----------

func TestFailedStatusTracking(t *testing.T) {
	// Trigger some failures
	fasttp.NewFunc(t, "fail 404", "/items/99999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
	fasttp.NewFunc(t, "fail 400", "/items/abc", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusBadRequest,
			JSON:   true,
		}
	})
	fasttp.NewFunc(t, "fail no auth", "/auth/profile", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusUnauthorized,
			JSON:   true,
		}
	})

	// Verify they were tracked (by admin)
	fasttp.NewFunc(t, "check failures", "/admin/failed-statuses", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer admin",
			},
			Data: func(data interface{}) interface{} {
				m := data.(map[string]interface{})
				failures := m["failures"].([]interface{})
				// Find our specific failures
				has404 := false
				has400 := false
				has401 := false
				for _, f := range failures {
					fm := f.(map[string]interface{})
					switch int(fm["status"].(float64)) {
					case 404:
						has404 = true
					case 400:
						has400 = true
					case 401:
						has401 = true
					}
				}
				assert.True(t, has404, "should track 404")
				assert.True(t, has400, "should track 400")
				assert.True(t, has401, "should track 401")
				return nil
			},
		}
	})
}
