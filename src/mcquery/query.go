package mcquery

import (
	"bufio"
	"encoding/binary"
	"strconv"
	"strings"
	"unicode"
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
func Handshake(rw *bufio.ReadWriter) (int32, error) {

	var response response
	request := request{
		Magic:     magicBytes(),
		Type:      handshakeType,
		SessionId: 1, // TODO figure this out
	}

	err := binary.Write(rw, binary.BigEndian, request)
	if err != nil {
		return 0, err
	}
	if err = rw.Flush(); err != nil {
		return 0, err
	}
	if err = binary.Read(rw, binary.BigEndian, &response); err != nil {
		return 0, err
	}

	payload, err := rw.ReadString(0)
	if err != nil {
		return 0, err
	}

	payload = strings.TrimRightFunc(payload, func(r rune) bool {
		return !unicode.IsDigit(r)
	})

	challenge, err := strconv.ParseInt(payload, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(challenge), nil
}
