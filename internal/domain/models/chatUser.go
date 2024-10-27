package models

import (
	"github.com/google/uuid"
	pb "github.com/yervsil/grpc-chat/pkg/api/chat"
)

type ChatUser struct {
	ID   	 	  uuid.UUID
	Name 	 	  string
	RoomName 	  string
	Message  	  chan *Message
	MessageStream pb.ChatService_ChatServer
}