package pkg

type LoadMode int

func (loadMode LoadMode) NeedMode(mode LoadMode) bool {
	return loadMode&mode == mode
}

const (
	LoadProvider LoadMode = iota
	LoadInjector
)

type LoadConfig struct {
	LoadMode LoadMode
}
