package chat

import "github.com/yervsil/grpc-chat/internal/domain/models"

type Room struct {
	Name   string
	users  map[string]*models.ChatUser
	closed bool
}

func NewRoom(name string) *Room {
	return &Room{
		Name:  name,
		users: make(map[string]*models.ChatUser),
	}
}