package breaker

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"sre-breaker/breaker/utils"
	"strings"

	"github.com/gin-gonic/gin"

	"google.golang.org/grpc"
)

const breakerSeparator = "://"

// BreakerHandler returns a break circuit middleware.
func GinBreakerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		brk := NewBreaker(
			WithName(
				strings.Join([]string{c.Request.Method, c.Request.URL.Path},
					breakerSeparator),
			),
		)

		promise, err := brk.Allow()
		if err != nil {
			log.Printf("[http] dropped, %s - %s", c.Request.RequestURI, c.Request.UserAgent())
			c.Writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		c.Next()

		repCode := c.Writer.Status()
		if repCode < http.StatusInternalServerError {
			promise.Accept()
		} else {
			promise.Reject(fmt.Sprintf("%d %s", repCode, http.StatusText(repCode)))
		}
	}
}

// BreakerInterceptor is an interceptor that acts as a circuit breaker.
func GrpcBreakerInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	breakerName := path.Join(cc.Target(), method)
	return DoWithAcceptable(breakerName, func() error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}, utils.Acceptable)
}
