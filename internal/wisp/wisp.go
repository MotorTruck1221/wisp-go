package wisp

import (
    "net/http"
    "fmt"
    "encoding/binary"
    "bytes"
    "github.com/gobwas/ws/wsutil"
    "github.com/gobwas/ws"
    "net"
    "sync"
    "C"
    "syscall"
)

const (
    connectType = 0x01
    dataType = 0x02
    continueType = 0x03
    closeType = 0x04
    tcpType = 0x01
    udpType = 0x02
    //huge max size for now
    maxBufferSize = 128
)

type WispPacket struct {
    Type uint8
    StreamID uint32
    Payload []byte
}

type ConnectPacket struct {
    Type uint8
    DestinationPort uint16
    DestinationHostname []byte
}

type DataPacket struct {
    StreamPayload []byte
}

type DataResponsePacket struct {
    StreamPayload []byte 
}

type ContinuePacket struct {
    BufferRemaining uint32
}

type ClosePacket struct {
    CloseType uint8
    StreamID uint32
    Reason uint8
}

type WispConnection struct {
    StreamID uint32 
    Conn net.Conn 
    BufferRemaining uint32 
}

//create a map to store all the tcp connections utilizing the WispConnection struct
var connections = make(map[uint32]*WispConnection)

func readPacket(conn net.Conn, channel chan WispPacket) {
    packet := WispPacket{}
    defer close(channel)
    for {
        msg, op, err := wsutil.ReadClientData(conn)
        if err != nil {
            fmt.Println("Error reading message! (channel may be closed): ", err)
            return
        }
        //fmt.Println("Received message: ", msg)
        fmt.Println("Received opcode: ", op)
        if op == ws.OpBinary {
            packet.Type = msg[0]
            packet.StreamID = binary.LittleEndian.Uint32(msg[1:5])
            packet.Payload = msg[5:]
            //fmt.Println("Received packet: ", packet.Type, packet.StreamID, packet.Payload)
            channel <- packet
        }
    }
}

func continuePacket(streamID uint32, bufferRemaining uint32, conn net.Conn) {
    packet := ContinuePacket{BufferRemaining: bufferRemaining}
    buffer := new(bytes.Buffer)
    binary.Write(buffer, binary.LittleEndian, packet.BufferRemaining)
    //create the wisp packet 
    wispPacket := WispPacket{Type: continueType, StreamID: streamID, Payload: buffer.Bytes()}
    //send the wisp packet 
    wispBuffer := new(bytes.Buffer)
    binary.Write(wispBuffer, binary.LittleEndian, wispPacket.Type)
    binary.Write(wispBuffer, binary.LittleEndian, wispPacket.StreamID)
    wispBuffer.Write(wispPacket.Payload)
    fmt.Println("Sending continue packet: ", wispBuffer.Bytes())
    wsutil.WriteServerMessage(conn, ws.OpBinary, wispBuffer.Bytes())
}

func tcpHandler(port uint16, hostname string, streamID uint32, waitGroup *sync.WaitGroup, wsConn net.Conn) {
}



func closePacket(streamID uint32, conn net.Conn, reason uint8) {
    packet := ClosePacket{CloseType: closeType, StreamID: streamID, Reason: reason}
    buffer := new(bytes.Buffer)
    binary.Write(buffer, binary.LittleEndian, packet.CloseType)
    binary.Write(buffer, binary.LittleEndian, packet.StreamID)
    binary.Write(buffer, binary.LittleEndian, packet.Reason)
    fmt.Println("Sending close packet: ", buffer.Bytes())
    wispPacket := WispPacket{Type: closeType, StreamID: streamID, Payload: buffer.Bytes()}
    wispBuffer := new(bytes.Buffer)
    binary.Write(wispBuffer, binary.LittleEndian, wispPacket.Type)
    binary.Write(wispBuffer, binary.LittleEndian, wispPacket.StreamID)
    wispBuffer.Write(wispPacket.Payload)
    wsutil.WriteServerMessage(conn, ws.OpBinary, wispBuffer.Bytes())
}

func handlePacket(channel chan WispPacket, conn net.Conn) {
    defer conn.Close()
}

func wisp(w http.ResponseWriter, r *http.Request) {
    conn := HandleUpgrade(w, r)
    fmt.Println("Connection established")
    continuePacket(0, maxBufferSize, conn)
    channel := make(chan WispPacket)
    go readPacket(conn, channel)
    go handlePacket(channel, conn)
    //defer conn.Close()
}
