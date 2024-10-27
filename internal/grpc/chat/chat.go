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
			close(chatUser.Message)
			c.Unregister <- chatUser
		}()

		for {
        msg, err := chatUser.MessageStream.Recv()
		fmt.Println(c.rooms)
        if err == io.EOF {
			log.Printf("Received error message from %s: %s, %s", msg.Sender, msg.Content, err.Error())
            return nil
        }
        if err != nil {
			log.Printf("Received error message from %s: %s, %s", msg.Sender, msg.Content, err.Error())
            return err
        }
        log.Printf("Received message from %s: %s", msg.Sender, msg.Content)

		message := &models.Message{Content: msg.Content, FromName: msg.Sender, FromUUID: chatUser.ID, RoomName: chatUser.RoomName}

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

	message := &pb.Message{Content: msg.Content, Sender: msg.FromName}

	err := chatUser.MessageStream.Send(message)
	if err != nil {
		log.Printf("error sending message from %s to %s, %s", msg.FromName, chatUser.Name, err.Error())
		return err
	}

	log.Printf("message Sended from %s to %s", msg.FromName, chatUser.Name)


}
}


func(c *ChatAPI) Kernel(ctx context.Context) {
	for {
		fmt.Println("0")
		select {
		case user := <-c.Register:		
		fmt.Println("1")
			room, ok := c.rooms[user.RoomName]
			if !ok {
				var m map[string]*models.ChatUser = map[string]*models.ChatUser{user.Name: user}

				c.rooms[user.RoomName] = &Room{
						Name: user.RoomName,
						users: m,					
				}
				fmt.Println("in register", c.rooms)
				continue
			}

			cUser, ok := room.users[user.Name]
			if !ok {
				room.users[user.Name] = user	
				fmt.Println("in register", c.rooms)
				continue
			}
			fmt.Println("in register", c.rooms)
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
			fmt.Println("2")
			room := c.rooms[msg.RoomName]
			fmt.Println("in transmit start", room.users)
			for name, user := range room.users {
				if msg.FromName != name {
					user.Message <- msg
				}
			}
			fmt.Println("in transmit end", room.users)
		case <- ctx.Done():
			fmt.Println("99")
			return
		}
	}
}