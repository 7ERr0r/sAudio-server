package server

import (
    "log"
    "golang.org/x/net/websocket"
    "fmt"
    "encoding/json"
    "encoding/binary"
    "math"
)

type WebConnection struct {
    server *WebServer
    ws *websocket.Conn
    ipaddr string
    closec chan bool
    writec chan interface{}

    dumping bool
    dumptick int
}

func NewWebConnection(server *WebServer, ws *websocket.Conn)(this *WebConnection){
    this = new(WebConnection)
    this.closec = make(chan bool, 0)
    this.writec = make(chan interface{}, 128)
    
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
    err = this.Init()
    if err != nil {
        return
    }
    err = this.ServeRead()
    return
}
func (this *WebConnection) Init()(err error){
    for _, channel := range this.server.audioserver.channelmap {
        msg := map[string]interface{}{"action": "addchannel", "name":channel.name}
        this.SendMessage(msg)

        for client, _ := range channel.clientmap {
            msg := map[string]interface{}{"action": "addclient", "name":client.addr.String(), "channel": channel.name, "description": client.String()}
            this.SendMessage(msg)
        }
    }
    return
}
func (this *WebConnection) Write(payload interface{})(err error){
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
func (this *WebConnection) WriteSoft(payload interface{})(err error){
    select {
    case <-this.closec:
        err = fmt.Errorf("connection is closed")
    default:
        select {
        case this.writec <- payload:
        default:
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
            err = websocket.Message.Send(this.ws, payload)
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
        case "startdump":
            this.dumping = true
        case "stopdump":
            this.dumping = false
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
    err = this.Write(string(payload))
    return
}

func (this *WebConnection) HandleDump(channel *Channel, data []float32){
    if this.dumping {
        if this.dumptick % 8 == 0 {
            skip := 2
            frame_size := len(data)
            
            dump := make([]byte, 4+(frame_size/skip)*4)
            binary.LittleEndian.PutUint32(dump[0:], 1)

            adump := dump[4:]
            for i := 0; i<frame_size/(2*skip); i++ {
                binary.LittleEndian.PutUint32(adump[4*(i*2):], math.Float32bits(data[i*2*skip]))
                binary.LittleEndian.PutUint32(adump[4*(1+i*2):], math.Float32bits(data[i*2*skip+1]))
            }
            this.WriteSoft(dump)
        }
        this.dumptick++
    }
    
}
func (this *WebConnection) HandleEvent(event Event){
    if join, ok := event.(ClientJoinChannelEvent); ok {
        msg := map[string]interface{}{"action": "addclient", "name":join.client.addr.String(), "channel": join.client.channel.name, "description": join.client.String()}
        this.SendMessage(msg)
    }else if quit, ok := event.(ClientQuitChannelEvent); ok {
        msg := map[string]interface{}{"action": "removeclient", "name":quit.client.addr.String(), "channel": quit.client.channel.name}
        this.SendMessage(msg)
    }else if update, ok := event.(ClientUpdateEvent); ok {
        msg := map[string]interface{}{"action": "updateclient", "name":update.client.addr.String(), "channel": update.client.channel.name, "description": update.client.String()}
        this.SendMessage(msg)
    }

}