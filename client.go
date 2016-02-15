package sAudio-server


type AudioClient struct {
	server *AudioServer
	addr *net.UDPAddr
	closec chan bool
	ticksfromlastpacket int
	handler ClientHandler
}

func NewAudioClient(server *AudioServer, addr *net.UDPAddr)(this *AudioClient){
	this = new(AudioClient)
	this.server = server
	this.addr = addr
	this.closec = make(chan bool)
	this.Log("connected")
	return
}

func (this *AudioClient) Log(str string){
	log.Printf("[%s]: %s", this.addr, str)
}
func (this *AudioClient) Logf(str string, a ...interface{}){
    this.Log(fmt.Sprintf(str, a...))
}
func (this *AudioClient) Receive(data []byte)(err error){
	//this.Log("received data")
	this.ticksfromlastpacket = 0
	if this.handler == nil {
		if string(data) == "ihazo" {
			this.InitOutput()
		}else if string(data) == "ihazi" {
			this.InitInput()
		}else{
			this.Logf("unrecognized packet: %s", data)
		}
	}else{
		err = this.handler.Receive(data)
	}
	return
}
func (this *AudioClient) InitOutput(){
	this.handler = NewOutputHandler(this)
	go func(){
		err := this.handler.Serve()
		if err != nil {
			this.Logf("output serve error: %s", err.Error())
		}
	}()

}
func (this *AudioClient) InitInput(){
	this.handler = NewInputHandler(this)
	go func(){
		err := this.handler.Serve()
		if err != nil {
			this.Logf("input serve error: %s", err.Error())
		}
	}()
}

func (this *AudioClient) Tick(){
	this.ticksfromlastpacket++
	if this.ticksfromlastpacket > 5 {
		this.Close()
	}
}
func (this *AudioClient) Close(){
	select {
	case <-this.closec:
		return
	default:
		close(this.closec)
		this.server.RemoveClient(this)
		this.Log("Closed")
	}
}

func (this *AudioClient) Send(data []byte)(err error){
	select {
	case <-this.closec:
		err = fmt.Errorf("audio client is closed")
		return
	default:
	}
	//this.Log("sending")
	n, err := this.server.listener.WriteToUDP(data, this.addr)
	if err != nil {
		return
	}
	if n != len(data) {
		err = fmt.Errorf("short WriteToUDP: sent %d != %d", n, len(data))
	}
	return
}
func (this *AudioClient) HandleBroadcast(data []float32){
	if outputhandler, ok := this.handler.(*OutputHandler); ok {
		outputhandler.HandleBroadcast(data)
	}
}

