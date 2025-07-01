package client

type Client interface {
	Connect(string) int
}
