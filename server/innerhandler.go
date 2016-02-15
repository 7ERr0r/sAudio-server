package server

type InnerHandler interface {
	HandleAudio(string, []float32)
	

}