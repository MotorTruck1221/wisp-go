package wisp

import (
    "net/http"
    "fmt"
    "github.com/gorilla/websocket"
)

const (
    connect = 0x01
    data = 0x02
    cont = 0x03
    close = 0x04
    tcpType = 0x01
    udpType = 0x02
)

var connections = make(map [string] *websocket.Conn)

type WispPacket struct {
    Type byte
    streamID uint32
    data []byte
}

func wisp(w http.ResponseWriter, r *http.Request) {
    ws, err := HandleUpgrade(w, r)
    if err != nil { fmt.Println("Error with upgrade: ", err) }
    defer ws.Close()
}
