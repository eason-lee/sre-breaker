package example

import (
	"context"
	"fmt"
	"sre-breaker/breaker"
	"time"

	"google.golang.org/grpc"
)

// NewClient ...
func NewClient(server string) (*grpc.ClientConn, error) {
	var cli *grpc.ClientConn
	options := []grpc.DialOption{grpc.WithChainUnaryInterceptor(breaker.GrpcBreakerInterceptor)}

	timeCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cli, err := grpc.DialContext(timeCtx, server, options...)
	if err != nil {
		return nil, fmt.Errorf("rpc dial: %s, error: %s, make sure rpc service %q is alread started",
			server, err.Error(), server)
	}

	return cli, nil

}
