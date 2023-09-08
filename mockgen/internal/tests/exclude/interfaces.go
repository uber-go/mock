package exclude

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=ignore -exclude_interfaces=IgnoreMe,IgnoreMe2

type IgnoreMe interface {
	A() bool
}

type IgnoreMe2 interface {
	~int
}

type GenerateMockForMe interface {
	B() int
}
