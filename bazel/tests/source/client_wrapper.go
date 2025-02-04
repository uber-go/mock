package client

type ClientWrapper interface {
	Client
	Close() error
}
