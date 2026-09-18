package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// DeprecatedAPI marks a legacy endpoint while keeping it functional.
func DeprecatedAPI(successor string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Warning", `299 - "Deprecated API; migrate to the v2 endpoint"`)
		if successor != "" {
			c.Header("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", successor))
		}
		c.Next()
	}
}
