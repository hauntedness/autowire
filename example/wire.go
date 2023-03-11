//go:build wireinject

package example

import (
	"github.com/google/wire"
	"github.com/huantedness/autowire/example/msg"
)

type Event struct {
	Message msg.Message
}

func NewEvent(message msg.Message) *Event {
	return &Event{
		Message: message,
	}
}
func InitEvent() *Event {
	panic(wire.Build(NewEvent, msg.NewMessage))
}
