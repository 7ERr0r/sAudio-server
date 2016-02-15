package server

import (
    "log"
    "golang.org/x/net/websocket"
    "fmt"
    "encoding/json"

)

type WebConnection struct {
    server *WebServer
    ws *websocket.Conn
    ipaddr string
    closec chan bool
    writec chan []byte
}

func NewWebConnection(server *WebServer, ws *websocket.Conn)(this *WebConnection){
    this = new(WebConnection)
    this.closec = make(chan bool, 0)
    this.writec = make(chan []byte, 128)
    
    this.server = server
    this.ws = ws
    this.ipaddr = ws.Request().RemoteAddr
    
    this.ws.PayloadType = websocket.TextFrame
    this.Log("Connected")
    return
}

func (this *WebConnection) Log(str string){
    log.Printf("[web%s]: %s\n", this.ipaddr, str)
    
}
func (this *WebConnection) Logf(str string, a ...interface{}){
    this.Log(fmt.Sprintf(str, a...))
}

func (this *WebConnection) Serve()(err error){
    go this.ServeWrite()
    err = this.ServeRead()
    return
}
func (this *WebConnection) Write(payload []byte)(err error){
    select {
    case <-this.closec:
        err = fmt.Errorf("connection is closed")
        return
    default:
        select {
        case this.writec <- payload:
            return
        case <-this.closec:
            err = fmt.Errorf("connection is closed")
            return
        }
    }
    return
}
func (this *WebConnection) ServeRead()(err error){
    var payload string
    var js interface{}
    for {
        err = websocket.Message.Receive(this.ws, &payload)
        if err != nil {
            return
        }
        this.Logf("received: %s", payload)
        err = json.Unmarshal([]byte(payload), &js)
        if object, ok := js.(map[string]interface{}); ok {
            err = this.HandleMessage(object)
            if err != nil {
                return
            }
        }else{
            err = fmt.Errorf("received json is not an object")
            return
        }
    }
    
    close(this.closec)
    return
}
func (this *WebConnection) ServeWrite()(err error){
    for {
        select {
        case payload := <-this.writec:
            err = websocket.Message.Send(this.ws, string(payload))
            if err != nil {
                return
            }

        case <-this.closec:
            return
        }
    }
    return
}

func (this *WebConnection) HandleMessage(js map[string]interface{})(err error){
    if action, ok := js["action"].(string); ok {
        switch(action){
        case "chat":
            if msg, ok := js["msg"].(string); ok {
                this.Logf("Chat: %v", msg)
                bc := map[string]interface{}{"action": "display","msg": msg}
                this.server.BroadcastMessage(bc)
            }
        }
    }else{


    }
    return
}
func (this *WebConnection) SendMessage(js interface{})(err error){
    payload, err := json.Marshal(js)
    if err != nil {
        return
    }
    err = this.Write(payload)
    return
}