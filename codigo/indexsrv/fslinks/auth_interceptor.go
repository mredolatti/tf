package fslinks

import (
	"context"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	errMissingUserID   = errors.New("no user id in context")
	errMissingServerID = errors.New("no server id in context")
)

type authInterceptor struct {
	userAccoounts repository.UserAccountRepository
}

func newAuthInterceptor(userAccoounts repository.UserAccountRepository) *authInterceptor {
	return &authInterceptor{userAccoounts: userAccoounts}
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

	serverID, ok := ctx.Value(ctxKeyServerID{}).(string)
	if !ok {
		return nil, errMissingServerID
	}

	token, err := a.getToken(ctx, userID, serverID)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", token), nil
}

func (a *authInterceptor) getToken(ctx context.Context, userID string, serverID string) (string, error) {
	acc, err := a.userAccoounts.Get(ctx, userID, serverID)
	if err != nil {
		return "", fmt.Errorf("error getting account from repository: %w", err)
	}

	// TODO(mredolatti): Validate token exp and re-fetch if necessary

	return acc.Token(), nil
}
