package msg

type Message interface{}

type message struct{}

// Text implements Message
func (*message) Text() string {
	panic("unimplemented")
}

func NewMessage() Message {
	return message{}
}
