package server

type Event interface {
	
}

type ClientJoinChannelEvent struct {
	client *AudioClient
	channel *Channel
}
type ClientQuitChannelEvent struct {
	client *AudioClient
	channel *Channel
}
type ClientUpdateEvent struct {
	client *AudioClient
}