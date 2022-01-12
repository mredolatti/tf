package main

import (
	"context"
	"fmt"
	"io"
	"time"

	usermgmt "github.com/mredolatti/tf/pocs/grpc"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:9638", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	client := usermgmt.NewUserManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	toCreate := map[string]int32{
		"pepe":  12,
		"juan":  21,
		"pedro": 123,
	}

	for k, v := range toCreate {
		r, err := client.CreateNewUser(ctx, &usermgmt.NewUser{Name: k, Age: v})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("Creado: ID: %d | Name: %s | Age: %d\n\n", r.GetId(), r.GetName(), r.GetAge())
	}

	fmt.Println("------------------------------")
	fmt.Println("")
	fmt.Println("")

	stream, err := client.ListUsers(context.Background(), &usermgmt.Empty{})
	if err != nil {
		panic(err.Error())
	}

	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("llego: id=%d name=%s, age=%d\n", user.GetId(), user.GetName(), user.GetAge())
	}
}
