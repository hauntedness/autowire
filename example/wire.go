//go:build wireinject

package example

import (
	"github.com/google/wire"
	"github.com/huantedness/autowire/example/msg"
)

type Event struct {
	Message *msg.Message
}

func NewEvent(message *msg.Message) *Event {
	return &Event{
		Message: message,
	}
}

func InitEventWithError() (*Event, error) {
	panic(wire.Build(NewEvent, msg.NewMessage))
}

func InitEvent() *Event {
	panic(wire.Build(NewEvent, msg.NewMessage))
}

func InitEventWithDummyReturn() *Event {
	wire.Build(NewEvent, msg.NewMessage)
	return nil
}

func InitEventWithWrongReturn() (*Event, func()) {
	wire.Build(NewEvent, msg.NewMessage)
	return nil, nil
}
