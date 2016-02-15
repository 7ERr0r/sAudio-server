package server



import (
	"log"
	"sync"
	"net/http"
	"net"
	"golang.org/x/net/websocket"
)




type WebServer struct {
	bindaddr string
    httpserver *http.Server
    audioserver *AudioServer
    listener net.Listener

    connections map[*WebConnection]bool
    connectionsmutex *sync.RWMutex


}

func NewWebServer(bindaddr string, audioserver *AudioServer)(this *WebServer){
    this = new(WebServer)
    this.bindaddr = bindaddr
    this.audioserver = audioserver


    this.connectionsmutex = new(sync.RWMutex)
    this.connections = make(map[*WebConnection]bool)
    mux := http.NewServeMux()
    //mux.Handle("/", xxx) // TODO
    mux.Handle("/socket/", websocket.Handler(this.ServeSocket))
    

    this.httpserver = &http.Server{
        Addr:           bindaddr,
        Handler:        mux,
        MaxHeaderBytes: 1 << 19,
    }
    
    return
}
func (this *WebServer) Serve()(err error){
	this.listener, err = net.Listen("tcp", this.bindaddr)
	if err != nil {
		return
	}
    log.Printf("Webserver listening on %s", this.httpserver.Addr)
    err = this.httpserver.Serve(this.listener)
    if err != nil {
        return
    }
    return
}



func (this *WebServer) AddConnection(conn *WebConnection){
    this.connectionsmutex.Lock()
    this.connections[conn] = true
    this.connectionsmutex.Unlock()
}
func (this *WebServer) RemoveConnection(conn *WebConnection){
    this.connectionsmutex.Lock()
    delete(this.connections, conn)
    this.connectionsmutex.Unlock()
}
func (this *WebServer) ServeSocket(ws *websocket.Conn){
    conn := NewWebConnection(this, ws)
    this.AddConnection(conn)
    err := conn.Serve()
    if err != nil {
        conn.Logf("ServeError: %s", err.Error())
    }else{
    	conn.Logf("End of serving")
	}
    this.RemoveConnection(conn)
}

func (this *WebServer) BroadcastMessage(js interface{}){
    for conn, _ := range this.connections {
        conn.SendMessage(js)
    }

}

func (this *WebServer) Close(){
	this.listener.Close()

}
