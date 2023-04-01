package server

import (
	"context"

	"google.golang.org/grpc"
)

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serverStreamWrapper) Context() context.Context { return w.ctx }

func wrapServerStream(ss grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &serverStreamWrapper{ss, ctx}
}
