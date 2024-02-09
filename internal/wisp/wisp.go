package wisp

import (
    "net/http"
    "fmt"
    "github.com/gorilla/websocket"
    "sync"
    "encoding/binary"
    "net"
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

type ContinuePacket struct {
    bufferRemaining uint32 
}

func packetParser(data []byte) WispPacket {
    dataType := data[0]
    streamID := binary.LittleEndian.Uint32(data[1:5])
    payload := data[5:]
    return WispPacket{dataType, streamID, payload}
}

func wsHandler(ws *websocket.Conn, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        _, data, err := ws.ReadMessage()
        if err != nil { fmt.Println("Error reading message:", err) }
        packet := packetParser(data)
        if packet.Type == connect {
            port := binary.LittleEndian.Uint16(packet.data[1:3])
            hostname := string(packet.data[3:])
            fmt.Println("Connecting to host:", hostname, "on port:", port)
            conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
            if err != nil { fmt.Println("Error connecting to host:", err) }
            fmt.Println("Connected to host", "conn:", conn)
        } else { fmt.Println("Unknown packet type received") }
    }
}

func wisp(w http.ResponseWriter, r *http.Request) {
    ws, err := HandleUpgrade(w, r)
    if err != nil { fmt.Println("Error with upgrade: ", err) }
    defer ws.Close()

    continuePacket := ContinuePacket{0}
    continuePacketBytes := make([]byte, 5)
    binary.LittleEndian.PutUint32(continuePacketBytes, continuePacket.bufferRemaining)
    ws.WriteMessage(websocket.BinaryMessage, continuePacketBytes)

    var wg sync.WaitGroup
    wg.Add(1)
    go wsHandler(ws, &wg)
    wg.Wait()
}
