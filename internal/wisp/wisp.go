package wisp

import (
    "net/http"
    "fmt"
    "github.com/gorilla/websocket"
    "sync"
    "encoding/binary"
    "net"
    "bufio"
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
    client *net.Conn
    remainingBuffer []byte
}

type ContinuePacket struct {
    bufferRemaining uint32 
}

type DataPacket struct {
    conmType []byte 
    streamID uint32
    data []byte
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
            conntype := packet.data[0]
            switch conntype {
                case tcpType:
                    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
                    if err != nil { fmt.Println("Error connecting to tcp server:", err) }
                    defer conn.Close()
                    connections[packet.streamID] = &Connection{&conn, make([]byte, 127)}
                    fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
                    reader, _ := bufio.NewReader(conn).Read(connections[packet.streamID].remainingBuffer)
                    fmt.Println("Data read from tcp server:", connections[packet.streamID].remainingBuffer)
                    fmt.Println("Reader:", reader)
                case udpType:
                default:
            }
        }
        if packet.Type == dataType {
            stream := connections[packet.streamID]
            _, err := (*stream.client).Write(packet.data)
            fmt.Println("Data written to tcp server:", packet.data)
            if err != nil { fmt.Println("Error writing to tcp server:", err) }
            //make the buffer decrease by one each time 
            stream.remainingBuffer = stream.remainingBuffer[1:]
            if len(stream.remainingBuffer) == 0 {
                stream.remainingBuffer = make([]byte, 127)
                continuePacket := ContinuePacket{127}
                continuePacketBytes := make([]byte, 5)
                binary.LittleEndian.PutUint32(continuePacketBytes, continuePacket.bufferRemaining)
                ws.WriteMessage(websocket.BinaryMessage, continuePacketBytes)
            }
        }
        if packet.Type == closeType {
        }
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
