package sAudio-server

type ClientHandler interface {
	Receive([]byte) error
	Serve() error
}