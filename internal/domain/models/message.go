package models

import "github.com/google/uuid"

type Message struct {
	Content  string
	FromName string
	FromUUID uuid.UUID
	RoomName string
}