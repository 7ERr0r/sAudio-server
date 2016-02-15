package main
 
import (
    "log"
    "flag"
    "github.com/Szperak/sAudio-server/server"
)


func main() {
    audioaddr := flag.String("a", ":42381", "(udp) audio bind address")
    webaddr := flag.String("w", ":42381", "(tcp) webserver bind address")
    flag.Parse()
    audioserver := server.NewAudioServer(*audioaddr)
    webserver := server.NewWebServer(*webaddr, audioserver)

    errchan := make(chan error, 2)
    go func(){
        err := audioserver.Serve()
        if err != nil {
            log.Printf("Audio server error: %s", err.Error())
        }
        errchan <- err
    }()
    go func(){
        err := webserver.Serve()
        if err != nil {
            log.Printf("Web server error: %s", err.Error())
        }
        errchan <- err
    }()
    err := <-errchan
    log.Printf("Closing both servers: %s", err.Error())
    audioserver.Close()
    webserver.Close()


    
}





/* old code


func main() {
    //filename := flag.String("f", "stream.raw", "stream to write")
    bindaddr := flag.String("b", ":42381", "bind address")
    flag.Parse()
    server := NewAudioServer(*bindaddr)
    go func(){
        err := serveAudio(server, *filename)
        if err != nil {
            log.Printf("Error serving audio: %s", err.Error())
        }
    }()
    err := server.Serve()
    if err != nil {
        log.Printf("Error: %s\n", err.Error())
    }


    
}

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

*/
