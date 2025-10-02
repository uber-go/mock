package exclude

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=exclude -exclude_interfaces=IgnoreMe,IgnoreMe2
//go:generate mockgen -source=interfaces.go -destination=ignore/mock.go IgnoreMe

type IgnoreMe interface {
	A() bool
}

type IgnoreMe2 interface {
	~int
}

type GenerateMockForMe interface {
	B() int
}
