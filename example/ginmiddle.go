package example

import (
	"sre-breaker/breaker"

	"github.com/gin-gonic/gin"
)

func ginMiddleware() *gin.Engine {
	engine := gin.Default()
	engine.Use(breaker.GinBreakerHandler())
	return engine
}

func main()  {
	r := ginMiddleware()
	r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    r.Run()
}
