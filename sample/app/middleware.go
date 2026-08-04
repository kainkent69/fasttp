package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Custom")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// authMiddleware requires a valid Bearer token.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusUnauthorized, "missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		if token != "Bearer secret" && token != "Bearer admin" {
			recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusForbidden, "invalid token")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid token"})
			return
		}
		c.Set("token", token)
		c.Next()
	}
}

// adminOnly middleware checks for admin scope token.
func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Get("token")
		if token != "Bearer admin" {
			recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusForbidden, "admin scope required")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin scope required"})
			return
		}
		c.Next()
	}
}

// rateLimitMiddleware allows `limit` requests per client (by X-Client-ID header).
func rateLimitMiddleware(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := c.GetHeader("X-Client-ID")
		if client == "" {
			client = c.ClientIP()
		}
		if !checkRateLimit(client, limit) {
			recordFailed(c.Request.Method, c.Request.URL.Path, http.StatusTooManyRequests, "rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
