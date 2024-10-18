package grpctest

import (
	"AuthDB/internal/api/user"
	"context"
	"testing"
	"time"
	pb "AuthDB/pkg/user_v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// gRPC StartServer func test

func TestStartGRPCServer(t *testing.T) {
	port := ":50052"
	accessService := &user.AccessService{}

	go func() {
		err := user.StartGRPCServer(port, accessService)
		require.NoError(t, err)
	}()

	time.Sleep(time.Second * 1)

	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	// Test request
	req := &pb.AccessRequest{Token: "valid-token", RequiredRole: "admin"}
	resp, err := client.CheckAccess(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, true, resp.HasAccess)
	require.Equal(t, "Access granted", resp.Message)
}
