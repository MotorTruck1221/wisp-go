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
    "sync"
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
    StreamPayload []rune
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
    for {
        msg, op, err := wsutil.ReadClientData(conn)
        if err != nil {
            fmt.Println("Error reading message: ", err)
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
    //defer close(channel)
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

func tcpHandler(port uint16, hostname string, streamID uint32, waitGroup *sync.WaitGroup) {
    //attempt basic net.Dial (for non TLS connections)
    fmt.Println("Attempting to connect to host:", hostname, "on port:", port, "with streamID:", streamID)
    //attempt to connect via TLS 
    tlsOptions := tls.Config{
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    }
    tcpConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), &tlsOptions)
    if err != nil {
        conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
        if err != nil {
            fmt.Println("Error connecting to host: ", err)
            return
        }
        fmt.Println("Connected to host (non-TLS)")
        wispConnection := &WispConnection{StreamID: streamID, Conn: conn, BufferRemaining: maxBufferSize}
        connections[streamID] = wispConnection 
    } else {
        fmt.Println("Connected to host (TLS)")
        wispConnection := &WispConnection{StreamID: streamID, Conn: tcpConn, BufferRemaining: maxBufferSize}
        connections[streamID] = wispConnection
    }
    defer waitGroup.Done()
    //defer delete(connections, streamID)
    //defer conn.Close()
}

func sendDataToClient(conn net.Conn, streamID uint32, payload []byte) {
    //write the payload to the client 
    wsutil.WriteServerMessage(conn, ws.OpBinary, payload)
    //decrement the buffer size 
    connection := connections[streamID]
    connection.BufferRemaining -= 1
    fmt.Println("Buffer remaining: ", connection.BufferRemaining)
    //if the buffer size is zero, send a continue packet 
    if connections[streamID].BufferRemaining == 0 {
        continuePacket(streamID, maxBufferSize, conn)
        //delete the connection from the map 
        delete(connections, streamID)
    }
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
    defer close(channel)
    defer conn.Close()
    for {
        packet, ok := <-channel
        fmt.Println("Received packet type: ", packet.Type)
        if !ok {
            fmt.Println("Channel closed")
        }
        //fmt.Println("Handling packet: ", packet)
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
            //create a new waitgroup 
            var waitGroup sync.WaitGroup 
            waitGroup.Add(1)
            go tcpHandler(connectPacket.DestinationPort, string(connectPacket.DestinationHostname), packet.StreamID, &waitGroup)
            waitGroup.Wait()            
        case dataType:
            fmt.Println("Data packet")
            dataPacket := DataPacket{}
            dataPacket.StreamPayload = []rune(string(packet.Payload))
            //print all of the available connections
            fmt.Println("Available connections: ", connections)
            //send the payload to the appropriate connection 
            conn := connections[packet.StreamID]
            fmt.Println("Connection: ", conn)
            conn.Conn.Write([]byte(string(dataPacket.StreamPayload)))
            //read everything from the connection (headers, etc)
            buffer := make([]byte, maxBufferSize)
            n, err := conn.Conn.Read(buffer)
            if err != nil { 
                fmt.Println("Error reading from connection: ", err)
            }
            fmt.Println("Received data: ", string(buffer[:n]))
            sendDataToClient(conn.Conn, packet.StreamID, buffer[:n])
         case closeType:
            fmt.Println("Close packet")
            //get the reason for closing the connection 
            reason := packet.Payload[0]
            closePacket(packet.StreamID, conn, reason)
            fmt.Println("Client decided to terminate with reason: ", reason)
            //close the connection(s)
            conn.Close()
            //close the tcp connection 
            connections[packet.StreamID].Conn.Close()
            //delete the connection from the map 
            delete(connections, packet.StreamID)
        }
    }
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
