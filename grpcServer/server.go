package grpcServer

import (
	"auth-api/helpers"
	"auth-api/models"
	"auth-api/proto"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedUserServiceServer
}

func (s *Server) AuthUser(ctx context.Context, token *proto.Token) (*proto.Valid, error) {
	id, err := helpers.GetIdFromToken(token.Token, false)

	if err != nil {
		return &proto.Valid{Bool: 0}, nil
	}

	u := models.FindUserByIdString(id)

	if u.Username == "" {
		return &proto.Valid{Bool: 0}, nil
	}

	return &proto.Valid{Bool: 1}, nil
}

func StartServer() {
	lis, err := net.Listen("tcp", ":9000")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterUserServiceServer(grpcServer, &Server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
