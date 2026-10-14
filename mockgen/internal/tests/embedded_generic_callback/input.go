package mypackage

//go:generate mockgen -destination=mock_test.go -package=mypackage go.uber.org/mock/mockgen/internal/tests/embedded_generic_callback MyInterface

// Regression case from https://github.com/uber-go/mock/issues/271#issuecomment-3442087301.
type Callback[T any] func(item T)

type Base[T any] interface {
	OnEvent(callback Callback[T])
}

type MyStruct struct{}

type MyInterface interface {
	Base[MyStruct]
}
