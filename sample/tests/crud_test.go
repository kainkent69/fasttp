package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	fasttp "github.com/kainkent69/fasttp/tests"
)

// ---------- GET /items ----------

func TestGetItemsEmpty(t *testing.T) {
	fasttp.NewFunc(t, "list empty", "/items", nil, func(tr *fasttp.T) fasttp.TestInfo {
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

func TestGetItemsWithMultipleItems(t *testing.T) {
	// Seed two items
	seed := func(name string, price float64) {
		body := jsonMarshal(map[string]interface{}{"name": name, "price": price})
		fasttp.NewFunc(t, "seed "+name, "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPost,
				Status:  http.StatusCreated,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			}
		})
	}
	seed("alpha", 10.0)
	seed("beta", 20.0)

	fasttp.NewFunc(t, "list multiple", "/items", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				arr := data.([]interface{})
				assert.GreaterOrEqual(t, len(arr), 2)
			},
		}
	})
}

func TestGetItemsFilterByName(t *testing.T) {
	fasttp.NewFunc(t, "filter items", "/items?q=alpha", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				arr := data.([]interface{})
				for _, item := range arr {
					m := item.(map[string]interface{})
					assert.Equal(t, "alpha", m["name"])
				}
			},
		}
	})
}

// ---------- GET /items/:id ----------

func TestGetItemFound(t *testing.T) {
	fasttp.NewFunc(t, "get item", "/items/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				assertMapEqual(t, data, "name", "alpha")
			},
		}
	})
}

func TestGetItemNotFound(t *testing.T) {
	fasttp.NewFunc(t, "get missing", "/items/99999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
			Data: func(data interface{}) {
				assertMapEqual(t, data, "error", "item not found")
			},
		}
	})
}

func TestGetItemBadIDFormat(t *testing.T) {
	fasttp.NewFunc(t, "get bad id", "/items/abc", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusBadRequest,
			JSON:   true,
			Data: func(data interface{}) {
				assertMapEqual(t, data, "error", "invalid id")
			},
		}
	})
}

func TestGetItemZeroID(t *testing.T) {
	fasttp.NewFunc(t, "get zero id", "/items/0", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

func TestGetItemNegativeID(t *testing.T) {
	fasttp.NewFunc(t, "get negative id", "/items/-1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

// ---------- POST /items ----------

func TestCreateItemSuccess(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{
		"name":  "newitem",
		"price": 29.99,
		"tags":  []string{"tag1", "tag2"},
		"active": true,
	})
	fasttp.NewFunc(t, "create item", "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "name", "newitem")
				assertMapEqual(t, data, "price", float64(29.99))
				assertMapEqual(t, data, "active", true)
				m := data.(map[string]interface{})
				assert.NotZero(t, m["id"])
				tags := m["tags"].([]interface{})
				assert.Len(t, tags, 2)
			},
		}
	})
}

func TestCreateItemBadJSON(t *testing.T) {
	fasttp.NewFunc(t, "create bad json", "/items", []byte("not json"), func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusBadRequest,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "error", "invalid JSON")
			},
		}
	})
}

func TestCreateItemMissingName(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{"price": 10.0})
	fasttp.NewFunc(t, "create no name", "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusUnprocessableEntity,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "error", "name is required")
			},
		}
	})
}

func TestCreateItemEmptyBody(t *testing.T) {
	fasttp.NewFunc(t, "create empty body", "/items", []byte{}, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusBadRequest,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestCreateItemExtraFields(t *testing.T) {
	body := []byte(`{"name":"extra","price":5.0,"unknown_field":"should be ignored"}`)
	fasttp.NewFunc(t, "create extra fields", "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "name", "extra")
			},
		}
	})
}

// ---------- PUT /items/:id ----------

func TestPutItemFullReplace(t *testing.T) {
	body := []byte(`{"name":"replaced","price":99.99,"active":false}`)
	fasttp.NewFunc(t, "put replace", "/items/1", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "name", "replaced")
				assertMapEqual(t, data, "price", float64(99.99))
				assertMapEqual(t, data, "active", false)
			},
		}
	})
}

func TestPutItemNotFound(t *testing.T) {
	body := []byte(`{"name":"nope"}`)
	fasttp.NewFunc(t, "put missing", "/items/99999", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusNotFound,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "error", "item not found")
			},
		}
	})
}

func TestPutItemBadID(t *testing.T) {
	body := []byte(`{"name":"x"}`)
	fasttp.NewFunc(t, "put bad id", "/items/abc", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusBadRequest,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

// ---------- PATCH /items/:id ----------

func TestPatchItemPartial(t *testing.T) {
	body := []byte(`{"price":55.55}`)
	fasttp.NewFunc(t, "patch partial", "/items/1", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPatch,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assertMapEqual(t, data, "price", float64(55.55))
			},
		}
	})
}

func TestPatchItemNotFound(t *testing.T) {
	body := []byte(`{"name":"x"}`)
	fasttp.NewFunc(t, "patch missing", "/items/99999", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPatch,
			Status:  http.StatusNotFound,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestPatchItemTags(t *testing.T) {
	body := []byte(`{"tags":["a","b","c"]}`)
	fasttp.NewFunc(t, "patch tags", "/items/1", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPatch,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				tags := m["tags"].([]interface{})
				assert.Len(t, tags, 3)
			},
		}
	})
}

// ---------- DELETE /items/:id ----------

func TestDeleteItemSuccess(t *testing.T) {
	fasttp.NewFunc(t, "delete item", "/items/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusOK,
			JSON:   true,
			Data: func(data interface{}) {
				assertMapEqual(t, data, "deleted", true)
			},
		}
	})
}

func TestDeleteItemNotFound(t *testing.T) {
	fasttp.NewFunc(t, "delete missing", "/items/99999", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusNotFound,
			JSON:   true,
		}
	})
}

func TestDeleteItemBadID(t *testing.T) {
	fasttp.NewFunc(t, "delete bad id", "/items/abc", nil, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusBadRequest,
			JSON:   true,
		}
	})
}

// ---------- Lifecycle ----------

func TestItemFullLifecycle(t *testing.T) {
	resetStore()
	t.Run("step1_create", func(t *testing.T) {
		body := jsonMarshal(map[string]interface{}{"name": "lifecycle", "price": 1.0})
		fasttp.NewFunc(t, "create", "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPost,
				Status:  http.StatusCreated,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
				Data: func(data interface{}) {
					assertMapEqual(t, data, "name", "lifecycle")
				},
			}
		})
	})
	t.Run("step2_read", func(t *testing.T) {
		fasttp.NewFunc(t, "read", "/items/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodGet,
				Status: http.StatusOK,
				JSON:   true,
				Data: func(data interface{}) {
					assertMapEqual(t, data, "name", "lifecycle")
				},
			}
		})
	})
	t.Run("step3_patch", func(t *testing.T) {
		body := []byte(`{"price":99.99}`)
		fasttp.NewFunc(t, "patch", "/items/1", body, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method:  http.MethodPatch,
				Status:  http.StatusOK,
				JSON:    true,
				Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
				Data: func(data interface{}) {
					assertMapEqual(t, data, "price", float64(99.99))
				},
			}
		})
	})
	t.Run("step4_delete", func(t *testing.T) {
		fasttp.NewFunc(t, "delete", "/items/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodDelete,
				Status: http.StatusOK,
				JSON:   true,
			}
		})
	})
	t.Run("step5_verify_gone", func(t *testing.T) {
		fasttp.NewFunc(t, "verify", "/items/1", nil, func(tr *fasttp.T) fasttp.TestInfo {
			return fasttp.TestInfo{
				Method: http.MethodGet,
				Status: http.StatusNotFound,
				JSON:   true,
			}
		})
	})
}

// ---------- Unique items (conflict) ----------

func TestCreateUniqueItemSuccess(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{"name": "unique-one", "price": 5.0})
	fasttp.NewFunc(t, "unique ok", "/unique-items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestCreateUniqueItemConflict(t *testing.T) {
	body := jsonMarshal(map[string]interface{}{"name": "unique-one", "price": 10.0})
	fasttp.NewFunc(t, "unique conflict", "/unique-items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusConflict,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				assert.Contains(t, data.(map[string]interface{})["error"], "name already exists")
			},
		}
	})
}

// ---------- Struct comparison ----------

func TestCRUDStructComparison(t *testing.T) {
	// Create
	body := jsonMarshal(map[string]interface{}{
		"name": "structitem", "price": 12.34, "active": true,
	})
	fasttp.NewFunc(t, "struct create", "/items", body, func(tr *fasttp.T) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
			Data: func(data interface{}) {
				m := data.(map[string]interface{})
				assert.Equal(t, "structitem", m["name"])
				assert.Equal(t, float64(12.34), m["price"])
				assert.Equal(t, true, m["active"])
				assert.NotZero(t, m["id"])
			},
		}
	})
}
