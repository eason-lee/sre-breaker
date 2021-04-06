package example

import (
	"sre-breaker/breaker"

	"github.com/gin-gonic/gin"
)

func ginMiddleware() {
	engine := gin.Default()
	engine.Use(breaker.GinBreakerHandler())
}
