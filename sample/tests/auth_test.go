package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kainkent69/fasttp"
)

// ---------- Auth: no token ----------

func TestAuthNoToken(t *testing.T) {
	fasttp.NewFunc(t, "auth no token", "/auth/profile", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusUnauthorized,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "error", "missing Authorization header")
				return nil
			},
		}
	})
}

func TestAuthNoTokenSettings(t *testing.T) {
	fasttp.NewFunc(t, "auth no token settings", "/auth/settings", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusUnauthorized,
			JSON:   true,
		}
	})
}

// ---------- Auth: bad token ----------

func TestAuthBadToken(t *testing.T) {
	fasttp.NewFunc(t, "auth bad token", "/auth/profile", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusForbidden,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer wrong",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "error", "invalid token")
				return nil
			},
		}
	})
}

func TestAuthEmptyToken(t *testing.T) {
	fasttp.NewFunc(t, "auth empty token", "/auth/profile", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusUnauthorized,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "",
			},
		}
	})
}

// ---------- Auth: valid token ----------

func TestAuthValidToken(t *testing.T) {
	fasttp.NewFunc(t, "auth valid token", "/auth/profile", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer secret",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "user", "authenticated")
				return nil
			},
		}
	})
}

func TestAuthValidTokenSettings(t *testing.T) {
	fasttp.NewFunc(t, "auth settings", "/auth/settings", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer secret",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "theme", "dark")
				return nil
			},
		}
	})
}

// ---------- Admin: requires admin scope ----------

func TestAdminWithUserToken(t *testing.T) {
	fasttp.NewFunc(t, "admin user token", "/admin/dashboard", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusForbidden,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer secret",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "error", "admin scope required")
				return nil
			},
		}
	})
}

func TestAdminWithAdminToken(t *testing.T) {
	fasttp.NewFunc(t, "admin token", "/admin/dashboard", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"Authorization": "Bearer admin",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "admin", true)
				return nil
			},
		}
	})
}

// ---------- Admin: failed status tracking ----------

func TestAdminFailedStatuses(t *testing.T) {
	fasttp.NewFunc(t, "failed statuses", "/admin/failed-statuses", nil, func(tr *fasttp.T) fasttp.TestInfo {
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
				// Should have accumulated failures from earlier auth tests
				assert.NotZero(t, len(failures), "should have tracked failed statuses")
				return nil
			},
		}
	})
}

// ---------- Rate limiting ----------

func TestRateLimitAllowed(t *testing.T) {
	for i := 0; i < 3; i++ {
		fasttp.NewFunc(t, "rate ok", "/rate-limited/data", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodGet,
				Status: http.StatusOK,
				JSON:   true,
				Headers: fasttp.HeaderMap{
					"X-Client-ID": "test-client",
				},
				Data: func(data interface{}) interface{} {
					assertMapEqual(t, data, "data", "precious")
					return nil
				},
			}
		})
	}
}

func TestRateLimitExceeded(t *testing.T) {
	// 4th request from same client should hit 429
	fasttp.NewFunc(t, "rate exceeded", "/rate-limited/data", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusTooManyRequests,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"X-Client-ID": "test-client",
			},
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "error", "rate limit exceeded")
				return nil
			},
		}
	})
}

func TestRateLimitDifferentClient(t *testing.T) {
	// Different client should still be allowed
	fasttp.NewFunc(t, "rate other client", "/rate-limited/data", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Headers: fasttp.HeaderMap{
				"X-Client-ID": "other-client",
			},
		}
	})
}

// ---------- Cookie ----------

func TestSetCookie(t *testing.T) {
	fasttp.NewFunc(t, "set cookie", "/set-cookie", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) interface{} {
				assertMapEqual(t, data, "cookie_set", true)
				return nil
			},
		}
	})
}
