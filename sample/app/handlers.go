package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter builds the complete router with all endpoints.
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// ---------- CRUD ----------
	r.GET("/items", listItems)
	r.GET("/items/:id", getItem)
	r.POST("/items", createItem)
	r.PUT("/items/:id", updateItem)
	r.PATCH("/items/:id", patchItem)
	r.DELETE("/items/:id", deleteItem)

	// ---------- Content types ----------
	ct := r.Group("/content")
	{
		ct.GET("/json", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"type": "json", "nested": gin.H{"deep": true}})
		})
		ct.GET("/xml", func(c *gin.Context) {
			c.XML(http.StatusOK, gin.H{"type": "xml"})
		})
		ct.GET("/text", func(c *gin.Context) {
			c.String(http.StatusOK, "plain text response")
		})
		ct.GET("/html", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body>hello</body></html>"))
		})
		ct.GET("/bytes", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/octet-stream", []byte{0x00, 0x01, 0x02, 0xFF})
		})
	}

	// ---------- Echo ----------
	r.POST("/echo", echoBody)
	r.GET("/headers", echoHeaders)
	r.POST("/reflect", reflectFull)

	// ---------- Status codes ----------
	r.GET("/status/:code", returnStatus)

	// ---------- Redirects ----------
	r.GET("/redirect", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/items")
	})
	r.GET("/redirect-temp", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/items")
	})

	// ---------- Slow / delayed ----------
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(500 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})
	r.POST("/stream", streamDelayed)
	r.GET("/stream/collect", collectStreamed)

	// ---------- Methods ----------
	r.Any("/methods", handleMethods)

	// ---------- Conflict / rate limit ----------
	r.POST("/unique-items", createUniqueItem)
	rl := r.Group("/rate-limited")
	rl.Use(rateLimitMiddleware(3))
	{
		rl.GET("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "precious"})
		})
	}

	// ---------- Auth routes ----------
	auth := r.Group("/auth")
	auth.Use(authMiddleware())
	{
		auth.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"user": "authenticated"})
		})
		auth.GET("/settings", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"theme": "dark"})
		})
	}
	admin := r.Group("/admin")
	admin.Use(authMiddleware())
	admin.Use(adminOnly())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"admin": true})
		})
		admin.GET("/failed-statuses", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"failures": getFailedStatuses()})
		})
	}

	// ---------- Edge cases ----------
	r.GET("/empty", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})
	r.GET("/empty-body", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.GET("/unicode", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "こんにちは世界", "emoji": "🚀✨"})
	})
	r.GET("/large", func(c *gin.Context) {
		data := make([]gin.H, 100)
		for i := range data {
			data[i] = gin.H{"index": i, "value": "abcdefghijklmnopqrstuvwxyz"}
		}
		c.JSON(http.StatusOK, gin.H{"items": data})
	})
	r.GET("/query-echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"q":    c.Query("q"),
			"page": c.Query("page"),
			"sort": c.QueryArray("sort"),
			"all":  c.Request.URL.Query(),
		})
	})
	r.GET("/set-cookie", func(c *gin.Context) {
		c.SetCookie("session", "abc123", 3600, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"cookie_set": true})
	})

	return r
}

// ---------- CRUD handlers ----------

func listItems(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	q := c.Query("q")
	activeOnly := c.Query("active") == "true"

	result := make([]Item, 0)
	for _, it := range items {
		if q != "" && it.Name != q {
			continue
		}
		if activeOnly && !it.Active {
			continue
		}
		result = append(result, it)
	}
	c.JSON(http.StatusOK, result)
}

func getItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	mu.Lock()
	it, ok := items[id]
	mu.Unlock()
	if !ok {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusNotFound, "item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, it)
}

func createItem(c *gin.Context) {
	var it Item
	if err := c.ShouldBindJSON(&it); err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if it.Name == "" {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusUnprocessableEntity, "name is required")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "name is required"})
		return
	}
	mu.Lock()
	it.ID = nextID
	nextID++
	if it.Tags == nil {
		it.Tags = []string{}
	}
	items[it.ID] = it
	mu.Unlock()
	c.JSON(http.StatusCreated, it)
}

func updateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var update Item
	if err := c.ShouldBindJSON(&update); err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	mu.Lock()
	_, ok := items[id]
	if !ok {
		mu.Unlock()
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusNotFound, "item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	// Full replacement
	update.ID = id
	if update.Tags == nil {
		update.Tags = []string{}
	}
	items[id] = update
	mu.Unlock()
	c.JSON(http.StatusOK, update)
}

func patchItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	mu.Lock()
	existing, ok := items[id]
	if !ok {
		mu.Unlock()
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusNotFound, "item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if v, ok := patch["name"]; ok {
		existing.Name = v.(string)
	}
	if v, ok := patch["price"]; ok {
		existing.Price = v.(float64)
	}
	if v, ok := patch["active"]; ok {
		existing.Active = v.(bool)
	}
	if v, ok := patch["tags"]; ok {
		tags := v.([]interface{})
		existing.Tags = make([]string, len(tags))
		for i, t := range tags {
			existing.Tags[i] = t.(string)
		}
	}
	items[id] = existing
	mu.Unlock()
	c.JSON(http.StatusOK, existing)
}

func deleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusBadRequest, "invalid id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	mu.Lock()
	_, ok := items[id]
	if ok {
		delete(items, id)
	}
	mu.Unlock()
	if !ok {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusNotFound, "item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ---------- Content-type handlers ----------

func echoHeaders(c *gin.Context) {
	hdrs := make(map[string]string)
	for k := range c.Request.Header {
		hdrs[k] = c.GetHeader(k)
	}
	c.JSON(http.StatusOK, hdrs)
}

func echoBody(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "invalid JSON")
		return
	}
	c.JSON(http.StatusOK, body)
}

func reflectFull(c *gin.Context) {
	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"method":  c.Request.Method,
		"path":    c.Request.URL.Path,
		"headers": c.Request.Header,
		"query":   c.Request.URL.Query(),
		"body":    body,
	})
}

// ---------- Status ----------

func returnStatus(c *gin.Context) {
	code, err := strconv.Atoi(c.Param("code"))
	if err != nil || code < 100 || code > 599 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status code"})
		return
	}
	c.JSON(code, gin.H{"status": code})
}

// ---------- Streaming / delayed ----------

func streamDelayed(c *gin.Context) {
	var req struct {
		Data  string `json:"data"`
		Delay int    `json:"delay_ms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if req.Delay == 0 {
		req.Delay = 100
	}
	go func() {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
		mu.Lock()
		entry := DelayedEntry{
			ID:        len(delayedOutputs) + 1,
			Data:      req.Data,
			CreatedAt: time.Now(),
		}
		delayedOutputs = append(delayedOutputs, entry)
		mu.Unlock()
	}()
	c.JSON(http.StatusAccepted, gin.H{"queued": true, "delay_ms": req.Delay})
}

func collectStreamed(c *gin.Context) {
	entries := collectDelayed()
	c.JSON(http.StatusOK, gin.H{"entries": entries, "count": len(entries)})
}

// ---------- Methods ----------

func handleMethods(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	case http.MethodPost:
		c.JSON(http.StatusCreated, gin.H{"method": "POST"})
	case http.MethodPut:
		c.JSON(http.StatusOK, gin.H{"method": "PUT"})
	case http.MethodPatch:
		c.JSON(http.StatusOK, gin.H{"method": "PATCH"})
	case http.MethodDelete:
		c.JSON(http.StatusOK, gin.H{"method": "DELETE"})
	case http.MethodHead:
		c.Status(http.StatusOK)
	case http.MethodOptions:
		c.Status(http.StatusNoContent)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// ---------- Conflict ----------

func createUniqueItem(c *gin.Context) {
	var it Item
	if err := c.ShouldBindJSON(&it); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if isNameUsed(it.Name) {
		recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusConflict, "name already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "name already exists: " + it.Name})
		return
	}
	mu.Lock()
	it.ID = nextID
	nextID++
	if it.Tags == nil {
		it.Tags = []string{}
	}
	items[it.ID] = it
	mu.Unlock()
	c.JSON(http.StatusCreated, it)
}
