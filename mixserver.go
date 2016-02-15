package main
 
import (
    "log"
    "net"
    "time"
    "flag"
    "fmt"
    "os"
    "io"
    "bufio"
    opus "github.com/southskies/golang-opus"
    "encoding/binary"
    "math"
)


func main() {
    //filename := flag.String("f", "stream.raw", "stream to write")
    bindaddr := flag.String("b", ":42381", "bind address")
    flag.Parse()
    server := NewAudioServer(*bindaddr)
    /*go func(){
    	err := serveAudio(server, *filename)
    	if err != nil {
    		log.Printf("Error serving audio: %s", err.Error())
    	}
    }()*/
    err := server.Serve()
    if err != nil {
        log.Printf("Error: %s\n", err.Error())
    }


    
}


const CHANNELS = 2
const SAMPLE_RATE = 48000






func writeAudio(server *AudioServer, r io.Reader)(err error){

    const FRAME_SIZE_MS = 5
	const FRAME_SIZE = (CHANNELS * SAMPLE_RATE * FRAME_SIZE_MS) / 1000
    buf := make([]byte, FRAME_SIZE*4)
    
    log.Printf("frame_size: %d", FRAME_SIZE)
    log.Printf("frame_size_ms: %d", FRAME_SIZE_MS)
    
    second := int64(1*time.Second)

    interval := time.Duration((second*int64(FRAME_SIZE_MS))/1000)
    log.Printf("interval: %v", interval)
    playtime := time.Now()

    
    var n int

    
    fbuf := make([]float32, FRAME_SIZE)
    for {
    	
    	n, err = io.ReadFull(r, buf)
        if err != nil {
            return
        }
        playtime = playtime.Add(interval)
        wait := playtime.Sub(time.Now())
        //log.Printf("waiting: %v", wait)
        time.Sleep(wait)

    	floats := n/4
    	
    	for i := 0; i<floats; i++ {
    		fbuf[i] = 0.5*math.Float32frombits(binary.LittleEndian.Uint32(buf[i*4:(i*4)+4]))
    	}
    	err = server.WriteAudio(fbuf[:floats])
        if err != nil {
            return
        }
    }
}
func serveAudio(server *AudioServer, filename string)(err error){
	file, err := os.Open(filename)
    if err != nil {
        return
    }
    defer file.Close()
    err = writeAudio(server, bufio.NewReader(file))
    if err != nil {
        err = fmt.Errorf("writing audio: %s", err.Error())
    }
    return
}




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
type AudioClient struct {
	server *AudioServer
	addr *net.UDPAddr
	closec chan bool
	ticksfromlastpacket int
	handler ClientHandler
}
type ClientHandler interface {
	Receive([]byte) error
	Serve() error
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


type OutputHandler struct {
	sess *AudioClient
	inputc chan []float32
	
}
func NewOutputHandler(sess *AudioClient)(this *OutputHandler){
	this = new(OutputHandler)
	this.sess = sess
	this.inputc = make(chan []float32, 64)
	this.sess.Log("output handler")

	return
}
func (this *OutputHandler) Receive(data []byte)(err error){
	return
}
func (this *OutputHandler) HandleBroadcast(data []float32){
	select {
	case this.inputc <- data:
	default:
	}
}
func (this *OutputHandler) Serve()(err error){
	enc := new(opus.Encoder)
	err = enc.Init(48000, 2, opus.APPLICATION_AUDIO)
	if err != nil {
		err = fmt.Errorf("could not create opus encoder: %s", err.Error())
		return
	}


	buf := make([]byte, 4000)
	var n int
	var sequenceId uint32
	for rawdata := range this.inputc {
		binary.LittleEndian.PutUint32(buf, sequenceId)
		n, err = enc.EncodeFloat32(rawdata, buf[4:])
		if err != nil {
			return
		}
		//if rand.Float32() > 0.05 { // Packet loss sim

		err = this.sess.Send(buf[:n+4])
		if err != nil {
			return
		}

		//}
		sequenceId++
	}

	return
}

type InputHandler struct {
	sess *AudioClient
	outputc chan []float32
	dec *opus.Decoder
}
func NewInputHandler(sess *AudioClient)(this *InputHandler){
	this = new(InputHandler)
	this.sess = sess
	this.sess.Log("input handler")

	this.outputc = make(chan []float32, 64)
	
	var err error
	this.dec, err = opus.NewDecoder(SAMPLE_RATE, CHANNELS)
	if err != nil || this.dec == nil {
		this.sess.Logf("err creating opus decoder: %s", err.Error())
		return
	}
	return
}
func (this *InputHandler) Serve()(err error){
	
	return
}
func (this *InputHandler) Receive(data []byte)(err error){
	if len(data) <= 4 {
		return
	}
	sequenceId := binary.LittleEndian.Uint32(data)
	_ = sequenceId

	const FRAME_SIZE_MS = 5
	const FRAME_SIZE = (CHANNELS * SAMPLE_RATE * FRAME_SIZE_MS) / 1000
	rawdata := make([]float32, FRAME_SIZE)
	_, err = this.dec.DecodeFloat32(data[4:], rawdata)
	if err != nil {
		return
	}
	//this.sess.Logf("n: %d, fs: %d", n, FRAME_SIZE)
	//this.outputc <- rawdata[:n]
	this.sess.server.WriteAudio(rawdata)
	return
}