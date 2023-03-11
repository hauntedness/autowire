package msg

type Message struct{}

// Text implements Message
func (*Message) Text() string {
	panic("unimplemented")
}
func NewMessage() Message {
	return Message{}
}
