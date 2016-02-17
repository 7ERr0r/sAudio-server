package server

import (
	"sync"
)
type Channel struct {
	name string
	server *AudioServer
	clientmap map[*AudioClient]bool
	clientmapmutex *sync.RWMutex
}

func NewChannel(server *AudioServer, name string)(this *Channel){
	this = new(Channel)
	this.clientmapmutex = new(sync.RWMutex)
	this.server = server
	this.name = name
	this.clientmap = make(map[*AudioClient]bool)
	return
}

func (this *Channel) WriteAudio(data []float32)(err error){
	broadcast := make([]float32, len(data))
	copy(broadcast, data)
	// Now you will ask: WHY DO YOU COPY THE WHOLE ARRAY?
	// Because audio reader and writer below are in another gouroutines (threads)
	this.clientmapmutex.RLock()
	for client, _ := range this.clientmap {
		client.HandleBroadcast(broadcast)

	}
	this.clientmapmutex.RUnlock()

	this.server.HandleChannelAudio(this, data)
	
	return
}
func (this *Channel) AddClient(client *AudioClient){
	this.clientmapmutex.Lock()
	this.clientmap[client] = true
	this.clientmapmutex.Unlock()
	this.server.CallEvent(ClientJoinChannelEvent{client, this})
}
func (this *Channel) RemoveClient(client *AudioClient){
	this.clientmapmutex.Lock()
	delete(this.clientmap, client)
	this.clientmapmutex.Unlock()
	this.server.CallEvent(ClientQuitChannelEvent{client, this})
}
