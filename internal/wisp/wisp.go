package wisp

import (
	"encoding/json"
	"fmt"
	"net/http"
    "github.com/gorilla/websocket"
    "encoding/binary"
    "bytes"
    "net"
)

type Packet struct {
	PacketType uint8      `json:"packetType"`
	StreamID   uint32     `json:"streamID"`
	Payload    []byte     `json:"payload"`
}

type ConnectPayload struct {
	StreamType          uint8  `json:"streamType"`
	DestinationPort     uint16 `json:"destinationPort"`
	DestinationHostname string `json:"destinationHostname"`
}

type DataPayload struct {
	StreamPayload []byte `json:"streamPayload"`
}

type ContinuePayload struct {
	BufferRemaining uint32 `json:"bufferRemaining"`
}

type ClosePayload struct {
	Reason uint8 `json:"reason"`
}


// MessageHeader represents the fixed-length header of a message.
type MessageHeader struct {
	MessageType uint8  // 1 byte for message type
	PayloadSize uint32 // 4 bytes for payload size
}

// Message represents a complete message with a header and payload.
type Message struct {
	Header  MessageHeader
	Payload []byte
}

// EncodeMessage encodes a message into a byte slice.
func EncodeMessage(msgType uint8, payload []byte) ([]byte, error) {
	header := MessageHeader{
		MessageType: msgType,
		PayloadSize: uint32(len(payload)),
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}

	messageBytes := append(headerBytes, payload...)
	return messageBytes, nil
}

// DecodeMessage decodes a byte slice into a message.
func DecodeMessage(data []byte) (*Message, error) {
	var header MessageHeader
	headerBuf := bytes.NewReader(data[:5])
	err := binary.Read(headerBuf, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	payload := data[5:]
	return &Message{Header: header, Payload: payload}, nil
}

func wispHandler(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
        Subprotocols: []string{"wisp-v1"},
    }
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil { fmt.Println("Error upgrading to WebSocket:", err) }
	// Send the initial "continue" payload
	continuePayload := ContinuePayload{BufferRemaining: 0}
	payload, err := json.Marshal(continuePayload)
	if err != nil {
		fmt.Println("Error marshaling continue payload:", err)
		return
	}

    messageBytes, err := EncodeMessage(0, payload)
    if err != nil { fmt.Println("Error encoding message:", err) }
    
    ws.WriteMessage(websocket.BinaryMessage, messageBytes)
    if err != nil {
        fmt.Println("Error writing to WebSocket:", err)
    }

    for {
        _, messageBytes, err := ws.ReadMessage()
        if err != nil { fmt.Println("Error reading from WebSocket:", err) }
        message, err := DecodeMessage(messageBytes)
        if err != nil { fmt.Println("Error decoding message:", err) }
        switch message.Header.MessageType {
            case 1:
                fmt.Println("Received TCP connection request")
                conn, err := net.Dial("tcp", "www.google.com:443")
                c := make([]byte, 1024)
                if err != nil {
                    fmt.Println("Error dialing TCP connection:", err)
                }
                defer conn.Close()
                n, err := conn.Read(c)
                if err != nil {
                    fmt.Println("Error reading from TCP connection:", err)
                }
                fmt.Println("Received from TCP connection:", string(c[:n]))
                var dataPayload DataPayload 
                dataPayload.StreamPayload = c[:n]
                packet := Packet{PacketType: 1, StreamID: 0, Payload: dataPayload.StreamPayload}
                packetBytes, err := json.Marshal(packet)
                if err != nil { fmt.Println("Error marshaling packet:", err) }
                ws.WriteMessage(websocket.BinaryMessage, packetBytes)
            case 2:
                fmt.Println("Received UDP connection request")
        }
    }
}
