package server

type InnerHandler interface {
	HandleAudio(*Channel, []float32)
	HandleEvent(Event)

}