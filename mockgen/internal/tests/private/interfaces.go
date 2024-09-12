package private

//go:generate mockgen -typed -private -source=interfaces.go -destination=interfaces_mock.go -package=private

type HelloWorld interface {
	Hi() string
}
