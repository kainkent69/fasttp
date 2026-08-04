package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetupRouter builds the complete chi-REST-style router.
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "root.") })
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.GET("/panic", func(c *gin.Context) { panic("test panic") })

	// Articles sub-router group
	articles := r.Group("/articles")
	{
		articles.GET("", paginate(), ListArticles)
		articles.POST("", CreateArticle)
		articles.GET("/search", SearchArticles)
		articles.GET("/:idOrSlug", ArticleCtx, GetArticle)
		articles.PUT("/:idOrSlug", ArticleCtx, UpdateArticle)
		articles.DELETE("/:idOrSlug", ArticleCtx, DeleteArticle)
	}

	// Admin mounted router
	admin := r.Group("/admin")
	admin.Use(AdminOnly)
	{
		admin.GET("", func(c *gin.Context) { c.String(http.StatusOK, "admin: index") })
		admin.GET("/accounts", func(c *gin.Context) { c.String(http.StatusOK, "admin: list accounts..") })
		admin.GET("/users/:userID", func(c *gin.Context) {
			c.String(http.StatusOK, "admin: view user id "+c.Param("userID"))
		})
	}

	return r
}

// ---------- middleware ----------

// ArticleCtx loads an article by ID or slug and stores it on the Gin context.
// Returns 404 if not found, 400 if ID format is invalid but param is all-numeric.
func ArticleCtx(c *gin.Context) {
	idOrSlug := c.Param("idOrSlug")

	// Try numeric ID first
	if id, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		a, err := DBGetArticle(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.Set("article", a)
		c.Next()
		return
	}

	// Not numeric — try as slug
	a, err := DBGetArticleBySlug(idOrSlug)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.Set("article", a)
	c.Next()
}

// AdminOnly checks the X-Admin header to restrict access.
// In the chi example this reads from context value "acl.admin".
func AdminOnly(c *gin.Context) {
	if c.GetHeader("X-Admin") != "true" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	c.Next()
}

// paginate is a middleware stub for page/limit query params.
func paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Stub: parse ?page= & ?limit= from c.Query, set defaults
		c.Next()
	}
}

// ---------- handlers ----------

func ListArticles(c *gin.Context) {
	articles := DBListArticles()
	c.JSON(http.StatusOK, articles)
}

func CreateArticle(c *gin.Context) {
	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	a := DBCreateArticle(req.Title, 1) // default user ID 1
	c.JSON(http.StatusCreated, a)
}

func GetArticle(c *gin.Context) {
	a, exists := c.Get("article")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func UpdateArticle(c *gin.Context) {
	aRaw, exists := c.Get("article")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	a := aRaw.(Article)

	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	updated, err := DBUpdateArticle(a.ID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteArticle(c *gin.Context) {
	aRaw, exists := c.Get("article")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	a := aRaw.(Article)

	if err := DBDeleteArticle(a.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "id": a.ID})
}

func SearchArticles(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}
	results := DBSearchArticles(q)
	c.JSON(http.StatusOK, results)
}
