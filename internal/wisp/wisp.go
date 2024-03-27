package wisp

import (
    "net/http"
    "fmt"
    "encoding/binary"
    "bytes"
    "github.com/gobwas/ws/wsutil"
    "github.com/gobwas/ws"
    "net"
    "crypto/tls"
)

const (
    connectType = 0x01
    dataType = 0x02
    continueType = 0x03
    closeType = 0x04
    tcpType = 0x01
    udpType = 0x02
    //huge max size for now
    maxBufferSize = 1000000
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

type ContinuePacket struct {
    BufferRemaining uint32
}

type ClosePacket struct {
    CloseType uint8
    StreamID uint32
    Reason uint8
}

func readPacket(conn net.Conn, channel chan WispPacket) {
    packet := WispPacket{}
    msg, op, _ := wsutil.ReadClientData(conn)
    fmt.Println("Received message: ", msg)
    fmt.Println("Received opcode: ", op)
    if op == ws.OpBinary {
        packet.Type = msg[0]
        packet.StreamID = binary.LittleEndian.Uint32(msg[1:5])
        packet.Payload = msg[5:]
        fmt.Println("Received packet: ", packet.Type, packet.StreamID, packet.Payload)
        channel <- packet
    }
    //close(channel)
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

func tcpHandler(port uint16, hostname string, channel chan WispPacket) {
    //attempt basic net.Dial (for non TLS connections)
    fmt.Println("Attempting to connect to host: ", hostname , " on port: ", port)
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
    if err != nil {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true,
        }
        //attempt TLS connection if basic net.Dial fails
        conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), tlsConfig)
        if err != nil {
            fmt.Println("Failed to connect to host: ", err)
            return
        }
    }
    defer conn.Close()
    fmt.Println("Connected to host")
}

func handlePacket(channel chan WispPacket, conn net.Conn) {
    for {
        packet, ok := <-channel
        if !ok {
            fmt.Println("Channel closed")
            return
        }
        fmt.Println("Handling packet: ", packet)
        switch packet.Type {
        case connectType:
            fmt.Println("Connect packet")
            connectPacket := ConnectPacket{}
            //the port comes right after the destination hostname
            connectPacket.DestinationHostname = packet.Payload[3:]
            connectPacket.DestinationPort = binary.LittleEndian.Uint16(packet.Payload[1:3])
            //either TCP or UDP 
            connectPacket.Type = packet.Payload[0]
            fmt.Println("TCP or UDP: ", connectPacket.Type)
            fmt.Println("Destination port: ", connectPacket.DestinationPort)
            fmt.Println("Destination hostname: ", string(connectPacket.DestinationHostname))
            tcpChannel := make(chan WispPacket)
            go tcpHandler(connectPacket.DestinationPort, string(connectPacket.DestinationHostname), tcpChannel)
        }
    }
}

func wisp(w http.ResponseWriter, r *http.Request) {
    conn := HandleUpgrade(w, r)
    fmt.Println("Connection established")
    continuePacket(0, 128, conn)
    channel := make(chan WispPacket)
    go readPacket(conn, channel)
    go handlePacket(channel, conn)
}
