package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewClientInterceptor(serviceToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.Pairs("x-service-token", serviceToken)

		if jwt, ok := ctx.Value(JWTKey).(string); ok && jwt != "" {
			md.Set("authentication", jwt)
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}
