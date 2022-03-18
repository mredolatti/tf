package server

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authInterceptor struct{}

var (
	errNoMetadata      = errors.New("no metadata in context")
	errNoAuthorization = errors.New("no authorization in metadata")
)

func (a *authInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := a.validate(ctx); err != nil {
			return nil, fmt.Errorf("error validating token in incoming rpc: %w", err)
		}
		return handler(ctx, req)
	}
}

func (a *authInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := a.validate(ss.Context()); err != nil {
			return fmt.Errorf("error validating token in incoming rpc: %w", err)
		}
		return handler(srv, ss)
	}
}

func (a *authInterceptor) validate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errNoMetadata
	}

	values := md["authorization"]
	if len(values) == 0 {
	}

	token := values[0]
	fmt.Println("llego token: ", token)
	return nil
}
