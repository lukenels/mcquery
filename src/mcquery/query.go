package mcquery

import (
	"bufio"
	"encoding/binary"
)

const (
	handshakeType = 9
)

func magicBytes() [2]byte {
	return [2]byte{0xFE, 0xFD}
}

type request struct {
	Magic     [2]byte
	Type      byte
	SessionId int32
}

type response struct {
	Type      byte
	SessionId int32
}

// Handshakes with the minecraft server, returning the challenge token
func Handshake(rw *bufio.ReadWriter) int32 {

	var response response
	request := request{
		Magic:     magicBytes(),
		Type:      handshakeType,
		SessionId: 1, // TODO figure this out
	}

	err := binary.Write(rw, binary.BigEndian, request)
	if err != nil {
		panic(err)
	}
	if err = rw.Flush(); err != nil {
		panic(err)
	}
	if err = binary.Read(rw, binary.BigEndian, &response); err != nil {
		panic(err)
	}

	return 0
}
