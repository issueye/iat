package grpc

import (
	"fmt"
	"iat/internal/service"
	pb "iat/pkg/pb/service"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(port int, chatSvc *service.ChatService) *Server {
	s := grpc.NewServer()

	// Register ChatService
	pb.RegisterChatServiceServer(s, NewChatHandler(chatSvc))

	return &Server{
		grpcServer: s,
		port:       port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	fmt.Printf("gRPC Server listening on port %d\n", s.port)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
