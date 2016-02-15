package server

type ClientHandler interface {
	Receive([]byte) error
	Serve() error
}