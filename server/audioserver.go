package server

import (
	"net"
	"fmt"
	"time"
	"log"
)

type AudioServer struct {
	bindaddr string
	listener *net.UDPConn
	addrclientmap map[string]*AudioClient
	handlermap map[InnerHandler]bool
	channelmap map[string]*Channel
	closec chan bool
}
func NewAudioServer(bindaddr string)(this *AudioServer){
	this = new(AudioServer)
	this.bindaddr = bindaddr
	this.addrclientmap = make(map[string]*AudioClient)
	this.handlermap = make(map[InnerHandler]bool)
	this.channelmap = make(map[string]*Channel)
	this.closec = make(chan bool)

	this.CreateChannel("default")
	return
}
func (this *AudioServer) Serve()(err error){
	listenaddr, err := net.ResolveUDPAddr("udp", this.bindaddr)
	if err != nil {
        return
    }
    this.listener, err = net.ListenUDP("udp", listenaddr)
    if err != nil {
        return
    }
    log.Printf("Audioserver listening on %s", listenaddr)
    go this.Ticker()
    var addr *net.UDPAddr
    var n int
    buf := make([]byte, 1<<15)
    for {
    	n, addr, err = this.listener.ReadFromUDP(buf)
    	if err != nil {
    		return
    	}
    	this.HandleData(buf[:n], addr)
    	if err != nil {
    		err = fmt.Errorf("handling data: %s", err.Error())
    		return
    	}
    }


    return
}
func (this *AudioServer) Close(){
	select {
	case <-this.closec:
		return
	default:
		this.listener.Close()
		for _, client := range this.addrclientmap {
			client.Close()
		}
	}
}
func (this *AudioServer) Ticker(){
	ticker := time.NewTicker(1*time.Second)
	for _ = range ticker.C {
		this.Tick()
	}
}
func (this *AudioServer) Tick(){
	for _, client := range this.addrclientmap {
		client.Tick()
	}
}
func AddrString(addr *net.UDPAddr) string {
	return string(addr.IP)+string([]byte{byte(addr.Port), byte(addr.Port>>8)})
}
func (this *AudioServer) AddClient(client *AudioClient){
	this.addrclientmap[AddrString(client.addr)] = client
}
func (this *AudioServer) RemoveClient(client *AudioClient){
	delete(this.addrclientmap, AddrString(client.addr))
}
func (this *AudioServer) HandleData(data []byte, addr *net.UDPAddr){
	var client *AudioClient
	var ok bool
    if client, ok = this.addrclientmap[AddrString(addr)]; !ok {
		client = NewAudioClient(this, addr)
		this.AddClient(client)
   		client.SetChannel(this.channelmap["default"])
    }
    err := client.Receive(data)
    if err != nil {
    	client.Logf("Error receiving data: %s", err.Error())
    }
    return
}
func (this *AudioServer) HandleChannelAudio(channel *Channel, data []float32)(err error){
	for handler, _ := range this.handlermap {
		handler.HandleAudio(channel, data)
	}
	return
}

func (this *AudioServer) CallEvent(event Event){
	for handler, _ := range this.handlermap {
		handler.HandleEvent(event)
	}
}

func (this *AudioServer) RegisterHandler(innerhandler InnerHandler){
	this.handlermap[innerhandler] = true
}
func (this *AudioServer) RemoveHandler(innerhandler InnerHandler){
	delete(this.handlermap, innerhandler)
}
func (this *AudioServer) CreateChannel(name string){
	this.channelmap[name] = NewChannel(this, name)
}