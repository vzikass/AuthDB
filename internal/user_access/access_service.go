package useraccess

import (
	"AuthDB/internal/helper"
	pb "AuthDB/pkg/access_v1"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type AccessService struct {
	pb.UnimplementedAuthServiceServer
}

func Register(grpcServer *grpc.Server, service *AccessService) {
	pb.RegisterAuthServiceServer(grpcServer, service)
}

func (s *AccessService) CheckAccess(ctx context.Context, req *pb.AccessRequest) (*pb.AccessResponse, error) {
	user, err := helper.GetUserByToken(req.Token)
	if err != nil {
		return &pb.AccessResponse{
			HasAccess: false,
			Message:   "Invalid token",
		}, nil
	}

	if user.Role != req.RequiredRole {
		return &pb.AccessResponse{
			HasAccess: false,
			Message:   "Access denied",
		}, nil
	}

	return &pb.AccessResponse{
		HasAccess: true,
		Message:   "Access granted",
	}, nil
}

func StartGRPCServer(port string, accessService *AccessService) error{
	grpcServer := grpc.NewServer()

	Register(grpcServer, accessService)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := grpcServer.Serve(lis); err != nil{
		log.Fatalf("failed to serve: %v", err)
	}
	
	return nil
}
