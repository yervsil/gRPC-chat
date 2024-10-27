package main

import (
	"context"
	"log"
	"net"

	"github.com/yervsil/grpc-chat/internal/grpc/chat"
	pb "github.com/yervsil/grpc-chat/pkg/api/chat"
	"google.golang.org/grpc"
)


func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    chatApi := chat.NewChatAPI()
    pb.RegisterChatServiceServer(s, chatApi)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go chatApi.Kernel(ctx)

    log.Println("Server is running on port 50051...")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
