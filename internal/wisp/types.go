package wisp

type ConnectPayload struct {
    uint8_t type
    uint16_t port
    char hostname[]
}

type WispPacket struct {
    uint8_t type
    uint32_t streamId
    char payload[]
}
