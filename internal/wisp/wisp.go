package wisp

import (
    "net/http"
    "fmt"
    "github.com/gorilla/websocket"
    "sync"
    "encoding/binary"
    "encoding/json"
    //"net"
    "crypto/tls"
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
    client **tls.Conn
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

func tcpHandler(conn *tls.Conn, tcpType byte, streamID uint32, ws *websocket.Conn) {
    //make this NON blocking
    reader := bufio.NewReader(conn)
    for {
        data, err := reader.ReadBytes('\n')
        if err != nil {
            fmt.Println("Error reading from tcp server:", err)
            return
        }
        fmt.Println("Data from tcp server:", string(data))
        //build out data packet 
        dataPacket := DataPacket{[]byte{tcpType}, streamID, data}
        dataPacketBytes, err := json.Marshal(dataPacket)
        if err != nil { fmt.Println("Error marshalling data packet:", err) }
        //create wisp packet 
        wispPacket := WispPacket{dataType, streamID, dataPacketBytes}
        wispPacketBytes, err := json.Marshal(wispPacket)
        if err != nil { fmt.Println("Error marshalling wisp packet:", err) }
        //send wisp packet to client as binary message 
        ws.WriteMessage(websocket.BinaryMessage, wispPacketBytes)
    }
}

func wsHandler(ws *websocket.Conn, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        _, data, err := ws.ReadMessage()
        if err != nil { fmt.Println("Error reading message:", err) }
        packet := packetParser(data)
        fmt.Println("Packet type:", packet.Type)
        if packet.Type == connect {
            port := binary.LittleEndian.Uint16(packet.data[1:3])
            hostname := string(packet.data[3:])
            conntype := packet.data[0]
            switch conntype {
                case tcpType:
                    //enabletls
                    tlsConfig := &tls.Config{
                        InsecureSkipVerify: true,
                    }
                    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), tlsConfig)
                    if err != nil { fmt.Println("Error connecting to tcp server:", err) }
                    defer conn.Close()
                    fmt.Println("Connected to tcp server")
                    connections[packet.streamID] = &Connection{&conn, make([]byte, 128)}
                    go tcpHandler(conn, tcpType, packet.streamID, ws)
                case udpType:
                default:
            }
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
