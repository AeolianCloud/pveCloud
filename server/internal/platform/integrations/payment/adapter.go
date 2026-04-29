package payment

type Adapter interface {
	Configured() bool
}

type Stub struct{}

func NewStub() Stub {
	return Stub{}
}

func (Stub) Configured() bool {
	return false
}
