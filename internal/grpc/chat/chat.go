package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/yervsil/grpc-chat/internal/domain/models"
	pb "github.com/yervsil/grpc-chat/pkg/api/chat"
	"google.golang.org/grpc/metadata"
)

type ChatAPI struct {
	pb.UnimplementedChatServiceServer
	rooms           map[string]*Room
	Register   		chan *models.ChatUser
	Unregister 		chan *models.ChatUser
	Transmit  		chan *models.Message
}

func NewChatAPI() *ChatAPI {
	service := &ChatAPI{
		rooms:           make(map[string]*Room),
		Register: make(chan *models.ChatUser),
		Unregister: make(chan *models.ChatUser),
		Transmit: make(chan *models.Message),
	}

	return service
}

func(c *ChatAPI) Chat(stream pb.ChatService_ChatServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())

	roomName := md.Get("room")[0]
	username := md.Get("username")[0]
	if roomName == "" || username == "" {
		return errors.New("empty data for chat")
	}
	fmt.Println(roomName, username)
	user := &models.ChatUser{
			ID: uuid.New(),
			Name: username,
			RoomName: roomName,
			MessageStream: stream,
			Message: make(chan *models.Message),
	}

	msg := &models.Message{
		Content: fmt.Sprintf("user %s joined to room %s", user.Name, user.RoomName),
		FromName: user.Name,
		FromUUID: user.ID,
		RoomName: user.RoomName,
	}
	go c.readMessage(user)

	c.Register <- user
	c.Transmit <- msg

	return c.writeMessage(user)
}

func(c *ChatAPI) readMessage(chatUser *models.ChatUser) error{
		defer func() {
			c.Unregister <- chatUser
			close(chatUser.Message)
		}()

		for {
        msg, err := chatUser.MessageStream.Recv()

        if err == io.EOF {
            return nil
        }
        if err != nil {
			log.Printf("Received error message from %s: %s, %s", chatUser.Name, msg.Content, err.Error())
            return err
        }
        log.Printf("Received message from %s: %s", chatUser.Name, msg.Content)

		message := &models.Message{Content: msg.Content, FromName: chatUser.Name, FromUUID: chatUser.ID, RoomName: chatUser.RoomName}

		c.Transmit <- message

	}
}

func(c *ChatAPI) writeMessage(chatUser *models.ChatUser) error{
	for {
	msg, ok := <- chatUser.Message
	if !ok {
		log.Printf("connection was closed")
		return errors.New("connection was closed")
	} 

	message := &pb.MessageResponse{Content: msg.Content, FromName: msg.FromName}

	err := chatUser.MessageStream.Send(message)
	if err != nil {
		log.Printf("error sending message from %s to %s, %s", msg.FromName, chatUser.Name, err.Error())
		return err
	}

	log.Printf("message Sended to %s from %s", chatUser.Name, msg.FromName)

}
}


func(c *ChatAPI) Kernel(ctx context.Context) {
	for {
		select {
		case user := <-c.Register:		
			room, ok := c.rooms[user.RoomName]
			if !ok {
				var m map[string]*models.ChatUser = map[string]*models.ChatUser{user.Name: user}

				c.rooms[user.RoomName] = &Room{
						Name: user.RoomName,
						users: m,					
				}
				continue
			}

			cUser, ok := room.users[user.Name]
			if !ok {
				room.users[user.Name] = user	
				continue
			}
			cUser.MessageStream = user.MessageStream
			
		case u := <-c.Unregister:
			room, ok := c.rooms[u.RoomName]
			if ok {		
				delete(room.users, u.Name)
				if len(room.users) == 0 {
					delete(c.rooms, u.RoomName)
				}
			}
		case msg := <-c.Transmit:
			room := c.rooms[msg.RoomName]
			for name, user := range room.users {
				if msg.FromName != name {
					user.Message <- msg
				}
			}
		case <- ctx.Done():
			return
		}
	}
}