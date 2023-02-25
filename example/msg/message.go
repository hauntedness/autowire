package msg

type Message interface {
	Text() string
}

var _ Message = (*message)(nil)

type message struct{}

// Text implements Message
func (*message) Text() string {
	panic("unimplemented")
}

func NewMessage() Message {
	return &message{}
}
