package wisp

import (
    "net/http"
    "fmt"
    "github.com/gorilla/websocket"
    "sync"
    "encoding/binary"
    //"net"
    //"encoding/json"
    "crypto/tls"
    //"bufio"
    "bytes"
)

const (
    connect = 0x01
    dataType = 0x02
    cont = 0x03
    closeType = 0x04
    tcpType = 0x01
    udpType = 0x02
)

var connections = make(map [uint32] *Connection)

type WispPacket struct {
    Type byte
    streamID uint32
    data []byte
}

type Connection struct {
    client **tls.Conn
    remainingBuffer []byte
}

type ContinuePacket struct {
    streamID uint32
    bufferRemaining uint32 
}

type DataPacket struct {
    connType byte 
    streamID uint32
    data []byte
}

func packetParser(data []byte) WispPacket {
    dataType := data[0]
    streamID := binary.LittleEndian.Uint32(data[1:5])
    payload := data[5:]
    return WispPacket{dataType, streamID, payload}
}

//this is driving me insane
//func tcpHandler(conn *tls.Conn, tcpType byte, streamID uint32, ws *websocket.Conn) {
//    reader := bufio.NewReader(conn)
//    for {
 //       data, _ := reader.ReadBytes('\n')
  //      dataPacket := DataPacket{tcpType, streamID, data}
   //     dataPacketBytes := new(bytes.Buffer)
    //    binary.Write(dataPacketBytes, binary.LittleEndian, dataPacket)
    //    wispPacket := WispPacket{dataType, streamID, dataPacketBytes.Bytes()}
     //   wispPacketBytes := new(bytes.Buffer)
     //   binary.Write(wispPacketBytes, binary.LittleEndian, wispPacket)
     //   fmt.Println("Sending data packet:", wispPacketBytes)
     //   ws.WriteMessage(websocket.BinaryMessage, wispPacketBytes.Bytes())
   // }
//}

func wsHandler(ws *websocket.Conn, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        _, data, err := ws.ReadMessage()
        if err != nil { fmt.Println("Error reading message:", err) }
        packet := packetParser(data)
        fmt.Println("Raw data:", data)
        fmt.Println("Received packet:", packet)
        fmt.Println("Packet type:", packet.Type)
    }
}

func wisp(w http.ResponseWriter, r *http.Request) {
    ws, err := HandleUpgrade(w, r)
    if err != nil { fmt.Println("Error with upgrade: ", err) }
    defer ws.Close()
    
    //send the continue packet with streamID 0 and bufferRemaining 127
    continuePacket := ContinuePacket{0, 127}
    continuePacketBytes := new(bytes.Buffer)
    binary.Write(continuePacketBytes, binary.LittleEndian, continuePacket)
    fmt.Println("Sending continue packet:", continuePacketBytes.Bytes())
    ws.WriteMessage(websocket.BinaryMessage, continuePacketBytes.Bytes())

    var wg sync.WaitGroup
    wg.Add(1)
    go wsHandler(ws, &wg)
    wg.Wait()
}
