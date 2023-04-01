package fslinks

import (
	"context"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	errMissingUserID   = errors.New("no user id in context")
	errMissingServerID = errors.New("no server id in context")
)

type authInterceptor struct {
	reg registrar.Interface
}

func newAuthInterceptor(reg registrar.Interface) *authInterceptor {
	return &authInterceptor{reg: reg}
}

func (a *authInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		fmt.Printf("req type: %T\n", req)
		fmt.Printf("req: %+v\n", req)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (a *authInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx, err := a.attachToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("error attaching token: %w", err)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (a *authInterceptor) attachToken(ctx context.Context) (context.Context, error) {
	userID, ok := ctx.Value(ctxKeyUserID{}).(string)
	if !ok {
		return nil, errMissingUserID
	}

	orgName, ok := ctx.Value(ctxKeyOrgName{}).(string)
	if !ok {
		return nil, errMissingServerID
	}

	serverName, ok := ctx.Value(ctxKeyServerName{}).(string)
	if !ok {
		return nil, errMissingServerID
	}

	token, err := a.reg.GetValidToken(ctx, userID, orgName, serverName)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", token.Raw()), nil
}
