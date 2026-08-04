package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kainkent69/fasttp/sample2/app"
	fasttp "github.com/kainkent69/fasttp/tests"
)

func init() {
	fasttp.Start(app.SetupRouter())
}

// ---------- root / ping / panic ----------

func TestRoot(t *testing.T) {
	fasttp.NewFunc(t, "root", "/", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			Buffer: true,
			Data: func(data interface{}) {
				assert.Equal(t, "root.", string(data.([]byte)))
			},
		}
	})
}

func TestPing(t *testing.T) {
	fasttp.NewFunc(t, "ping", "/ping", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			Buffer: true,
			Data: func(data interface{}) {
				assert.Equal(t, "pong", string(data.([]byte)))
			},
		}
	})
}

// ---------- GET /articles ----------

func TestListArticles(t *testing.T) {
	fasttp.NewFunc(t, "list articles", "/articles", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				arr := data.([]interface{})
				assert.Len(t, arr, 5)
				first := arr[0].(map[string]interface{})
				assert.Equal(t, "First Post", first["title"])
				assert.Equal(t, "first-post", first["slug"])
			},
		}
	})
}

// ---------- POST /articles ----------

func TestCreateArticle(t *testing.T) {
	body := []byte(`{"title":"New Article"}`)
	fasttp.NewFunc(t, "create article", "/articles", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "New Article", m["title"])
				assert.Equal(t, "new-article", m["slug"])
				assert.NotZero(t, m["id"])
			},
		}
	})
}

func TestCreateArticleNoTitle(t *testing.T) {
	body := []byte(`{}`)
	fasttp.NewFunc(t, "create no title", "/articles", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusBadRequest,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "title is required", m["error"])
			},
		}
	})
}

func TestCreateArticleBadJSON(t *testing.T) {
	fasttp.NewFunc(t, "create bad json", "/articles", []byte("not json"), func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusBadRequest,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

// ---------- GET /articles/:articleID ----------

func TestGetArticleByID(t *testing.T) {
	app.ResetStore()
	fasttp.NewFunc(t, "get article by id", "/articles/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, float64(1), m["id"])
				assert.Equal(t, "First Post", m["title"])
			},
		}
	})
}

func TestGetArticleNotFound(t *testing.T) {
	fasttp.NewFunc(t, "get article not found", "/articles/99999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

func TestGetArticleBadSlug(t *testing.T) {
	// Non-numeric, non-existent slug → 404
	fasttp.NewFunc(t, "get bad slug", "/articles/not-a-real-slug-xyz", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

// ---------- GET /articles/:articleSlug ----------

func TestGetArticleBySlug(t *testing.T) {
	app.ResetStore()
	fasttp.NewFunc(t, "get by slug", "/articles/hello-world", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, float64(3), m["id"])
				assert.Equal(t, "Hello World", m["title"])
			},
		}
	})
}

func TestGetArticleBySlugNotFound(t *testing.T) {
	fasttp.NewFunc(t, "get slug not found", "/articles/nonexistent-slug", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

// ---------- PUT /articles/:articleID ----------

func TestUpdateArticle(t *testing.T) {
	app.ResetStore()
	body := []byte(`{"title":"Updated Title"}`)
	fasttp.NewFunc(t, "update article", "/articles/1", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "Updated Title", m["title"])
				assert.Equal(t, "updated-title", m["slug"])
			},
		}
	})
}

func TestUpdateArticleNotFound(t *testing.T) {
	body := []byte(`{"title":"Nope"}`)
	fasttp.NewFunc(t, "update not found", "/articles/99999", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusNotFound,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

// ---------- DELETE /articles/:articleID ----------

func TestDeleteArticle(t *testing.T) {
	app.ResetStore()
	fasttp.NewFunc(t, "delete article", "/articles/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, true, m["deleted"])
			},
		}
	})
	// Verify deleted
	fasttp.NewFunc(t, "verify deleted", "/articles/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

func TestDeleteArticleNotFound(t *testing.T) {
	fasttp.NewFunc(t, "delete not found", "/articles/99999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

// ---------- GET /articles/search ----------

func TestSearchArticles(t *testing.T) {
	app.ResetStore()
	fasttp.NewFunc(t, "search articles", "/articles/search?q=go", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				arr := data.([]interface{})
				assert.GreaterOrEqual(t, len(arr), 1)
				for _, item := range arr {
					m := item.(map[string]interface{})
					assert.Contains(t, []string{"Go Is Great"}, m["title"])
				}
			},
		}
	})
}

func TestSearchArticlesNoQuery(t *testing.T) {
	fasttp.NewFunc(t, "search no query", "/articles/search", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusBadRequest,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "query parameter q is required", m["error"])
			},
		}
	})
}

func TestSearchArticlesNoResults(t *testing.T) {
	fasttp.NewFunc(t, "search no results", "/articles/search?q=zzzzzz", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				arr := data.([]interface{})
				assert.Empty(t, arr)
			},
		}
	})
}

// ---------- Admin routes ----------

func TestAdminIndex(t *testing.T) {
	fasttp.NewFunc(t, "admin index", "/admin", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Buffer:  true,
			Headers: fasttp.HeaderMap{"X-Admin": "true"},
			Data: func(data interface{}) {
				assert.Equal(t, "admin: index", string(data.([]byte)))
			},
		}
	})
}

func TestAdminAccounts(t *testing.T) {
	fasttp.NewFunc(t, "admin accounts", "/admin/accounts", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Buffer:  true,
			Headers: fasttp.HeaderMap{"X-Admin": "true"},
			Data: func(data interface{}) {
				assert.Equal(t, "admin: list accounts..", string(data.([]byte)))
			},
		}
	})
}

func TestAdminUserView(t *testing.T) {
	fasttp.NewFunc(t, "admin user view", "/admin/users/42", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Buffer:  true,
			Headers: fasttp.HeaderMap{"X-Admin": "true"},
			Data: func(data interface{}) {
				assert.Contains(t, string(data.([]byte)), "admin: view user id 42")
			},
		}
	})
}

func TestAdminNoHeader(t *testing.T) {
	fasttp.NewFunc(t, "admin no header", "/admin", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusForbidden,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "admin access required", m["error"])
			},
		}
	})
}

func TestAdminWrongHeader(t *testing.T) {
	fasttp.NewFunc(t, "admin wrong header", "/admin", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodGet,
			Status:  http.StatusForbidden,
			JSON:    true,
			Headers: fasttp.HeaderMap{"X-Admin": "false"},
		}
	})
}

// ---------- Article lifecycle ----------

func TestArticleLifecycle(t *testing.T) {
	app.ResetStore()

	t.Run("create", func(t *testing.T) {
		body := []byte(`{"title":"Lifecycle Test"}`)
		fasttp.NewFunc(t, "create", "/articles", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPost,
				Status:  http.StatusCreated,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
				Data: func(data interface{}) {
					m := data.(map[string]interface{})
					assert.Equal(t, "Lifecycle Test", m["title"])
				},
			}
		})
	})

	t.Run("read", func(t *testing.T) {
		fasttp.NewFunc(t, "read", "/articles/6", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodGet,
				Status: http.StatusOK,
				JSON:   true,
				Data: func(data interface{}) {
					m := data.(map[string]interface{})
					assert.Equal(t, "Lifecycle Test", m["title"])
				},
			}
		})
	})

	t.Run("update", func(t *testing.T) {
		body := []byte(`{"title":"Lifecycle Updated"}`)
		fasttp.NewFunc(t, "update", "/articles/6", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPut,
				Status:  http.StatusOK,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
				Data: func(data interface{}) {
					m := data.(map[string]interface{})
					assert.Equal(t, "Lifecycle Updated", m["title"])
					assert.Equal(t, "lifecycle-updated", m["slug"])
				},
			}
		})
	})

	t.Run("delete", func(t *testing.T) {
		fasttp.NewFunc(t, "delete", "/articles/6", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodDelete,
				Status: http.StatusOK,
				JSON:   true,
			}
		})
	})

	t.Run("verify gone", func(t *testing.T) {
		fasttp.NewFunc(t, "verify gone", "/articles/6", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodGet,
				Status: http.StatusNotFound,
			}
		})
	})
}

// ---------- URL format / slug priority ----------

func TestGetBySlugPriority(t *testing.T) {
	// /articles/rest-apis should match slug "rest-apis" (article ID 5)
	fasttp.NewFunc(t, "slug match", "/articles/rest-apis", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, float64(5), m["id"])
				assert.Equal(t, "REST APIs", m["title"])
			},
		}
	})
}

// ---------- Create uses default user ----------

func TestCreateArticleDefaultUser(t *testing.T) {
	app.ResetStore()
	body := []byte(`{"title":"User Post"}`)
	fasttp.NewFunc(t, "create with user", "/articles", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, float64(1), m["user_id"], "default user should be 1")
			},
		}
	})
}
