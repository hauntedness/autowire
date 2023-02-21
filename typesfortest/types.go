package typesfortest

type Person interface {
	SayHi() string
}

type Bob struct{}

// SayHi implements Person
func (*Bob) SayHi() string {
	panic("unimplemented")
}

var _ Person = (*Bob)(nil)

var NewBobFunc = func() *Bob {
	return &Bob{}
}

var Anonymous = struct{ Name string }{"Bob"}
