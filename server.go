package sAudio-server


type AudioServer struct {
	bindaddr string
	listener *net.UDPConn
	addrclientmap map[string]*AudioClient

}
func NewAudioServer(bindaddr string)(this *AudioServer){
	this = new(AudioServer)
	this.bindaddr = bindaddr
	this.addrclientmap = make(map[string]*AudioClient)
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
    defer this.listener.Close()
    go this.Ticker()
    var addr *net.UDPAddr
    var n int
    buf := make([]byte, 1<<15)
    for {
    	n, addr, err = this.listener.ReadFromUDP(buf)
    	if err != nil {
    		return
    	}
    	err = this.HandleData(buf[:n], addr)
    	if err != nil {
    		err = fmt.Errorf("handling data: %s", err.Error())
    		return
    	}
    }


    return
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
func (this *AudioServer) HandleData(data []byte, addr *net.UDPAddr)(err error){
	var client *AudioClient
	var ok bool
    if client, ok = this.addrclientmap[AddrString(addr)]; !ok {
		client = NewAudioClient(this, addr)
		this.AddClient(client)
   		
    }
    err = client.Receive(data)
    return
}
func (this *AudioServer) WriteAudio(data []float32)(err error){
	broadcast := make([]float32, len(data))
	copy(broadcast, data)
	// Now you will ask: WHY DO YOU COPY THE WHOLE ARRAY?
	// Because audio reader and writer below are in another gouroutines (threads)
	for _, client := range this.addrclientmap {
		client.HandleBroadcast(broadcast)

	}
	return
}