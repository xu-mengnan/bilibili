package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets a restrictive baseline for the bundled web UI.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' https: data:; "+
				"connect-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'; "+
				"form-action 'self'")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}
