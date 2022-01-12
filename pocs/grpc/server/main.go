package main

import (
	"context"
	"math/rand"
	"net"
	"sync"

	usermgmt "github.com/mredolatti/tf/pocs/grpc"
	"google.golang.org/grpc"
)

// UserManagementServer sarasa
type UserManagementServer struct {
	usermgmt.UnimplementedUserManagementServer

	users []usermgmt.User
	mutex sync.RWMutex
}

// CreateNewUser sarasa
func (u *UserManagementServer) CreateNewUser(ctx context.Context, in *usermgmt.NewUser) (*usermgmt.User, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.users = append(u.users, usermgmt.User{
		Name: in.GetName(),
		Id:   int32(rand.Intn(1000)),
		Age:  in.GetAge(),
	})

	return &u.users[len(u.users)-1], nil
}

// ListUsers sarasa
func (u *UserManagementServer) ListUsers(in *usermgmt.Empty, stream usermgmt.UserManagement_ListUsersServer) error {
	u.mutex.RLock()
	for idx := range u.users {
		stream.Send(&u.users[idx])
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":9638")
	if err != nil {
		panic(err.Error())
	}

	server := grpc.NewServer()
	usermgmt.RegisterUserManagementServer(server, &UserManagementServer{})
	if err = server.Serve(lis); err != nil {
		panic(err.Error())
	}
}
