package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authInterceptor struct {
	oauth2Wrapper oauth2.Interface
	logger        log.Interface
}

var (
	errNoMetadata      = errors.New("no metadata in context")
	errNoAuthorization = errors.New("no authorization in metadata")
)

func newAuthInterceptor(logger log.Interface, oauth2Wrapper oauth2.Interface) *authInterceptor {
	return &authInterceptor{
		oauth2Wrapper: oauth2Wrapper,
		logger:        logger,
	}
}

func (a *authInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, err := a.validate(ctx); err != nil { // TODO(user validation)
			return nil, fmt.Errorf("error validating token in incoming rpc: %w", err)
		}
		return handler(ctx, req)
	}
}

func (a *authInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		user, err := a.validate(ss.Context())
		if err != nil {
			return fmt.Errorf("error validating token in incoming rpc: %w", err)
		}

		// TODO(mredolatti): mover esto a un package separado y usar una key para guardar el user
		return handler(srv, wrapServerStream(ss, context.WithValue(ss.Context(), "user", user)))
	}
}

func (a *authInterceptor) validate(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errNoMetadata
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	user, err := a.verifyJWT(values[0])
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	// TODO(mredolatti): validar que el subject del token este en los SAN del client cert

	return user, nil
}

func (a *authInterceptor) verifyJWT(token string) (user string, err error) {
	claims, err := a.oauth2Wrapper.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("error validating incoming jwt: %w", err)
	}

	return claims.Subject, nil
}

type tokenTag struct{}

type customServerStream struct {
	grpc.ServerStream
	token *jwt.StandardClaims
}

func (c *customServerStream) Context() context.Context {
	return context.WithValue(c.ServerStream.Context(), tokenTag{}, c.token)
}
