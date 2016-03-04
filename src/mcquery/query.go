package mcquery

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	handshakeType = 9
	statusType    = 0
)

func magicBytes() [2]byte {
	return [2]byte{0xFE, 0xFD}
}

type handshakeRequest struct {
	Magic     [2]byte
	Type      byte
	SessionId int32
}

type basicStatRequest struct {
	Magic          [2]byte
	Type           byte
	SessionId      int32
	ChallengeToken int32
}

type responseHeader struct {
	Type      byte
	SessionId int32
}

type BasicStatResponse struct {
	Type       byte
	SessionId  int32
	Motd       string
	Gametype   string
	Map        string
	NumPlayers string
	MaxPlayers string
	HostPort   uint16
	HostIp     string
}

func Connect(ip string, port uint) (*bufio.ReadWriter, error, chan<- bool) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err, nil
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	kill := make(chan bool)
	go func() {
		<-kill
		conn.Close()
	}()

	return rw, nil, kill
}

func BasicStat(rw *bufio.ReadWriter, challenge int32) (*BasicStatResponse, error) {

	var response BasicStatResponse
	done := make(chan error)
	request := basicStatRequest{
		Magic:          magicBytes(),
		Type:           statusType,
		SessionId:      1,
		ChallengeToken: challenge,
	}

	go func() {

		err := binary.Write(rw, binary.BigEndian, request)
		if err != nil {
			done <- err
			return
		}
		rw.Flush()

		var responseHeader responseHeader
		err = binary.Read(rw, binary.BigEndian, &responseHeader)
		if err != nil {
			done <- err
			return
		}

		// All of this code here is really ugly, makes me want to switch to
		// haskell and use monads...

		mesg, err := rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]

		response.Motd = string(mesg)

		mesg, err = rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]
		response.Gametype = string(mesg)

		mesg, err = rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]
		response.Map = string(mesg)

		mesg, err = rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]
		response.NumPlayers = string(mesg)

		mesg, err = rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]
		response.MaxPlayers = string(mesg)

		var wrappedPort struct {
			Port uint16
		}
		err = binary.Read(rw, binary.LittleEndian, &wrappedPort)
		if err != nil {
			done <- err
			return
		}
		response.HostPort = wrappedPort.Port

		mesg, err = rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		mesg = mesg[:len(mesg)-1]
		response.HostIp = string(mesg)

		done <- nil
	}()

	select {
	case <-time.After(2000 * time.Millisecond):
		return nil, errors.New("Timed out")
	case err := <-done:
		if err == nil {
			return &response, nil
		} else {
			return nil, err
		}
	}
}

// Handshakes with the minecraft server, returning the challenge token
func Handshake(rw *bufio.ReadWriter) (int32, error) {

	var response responseHeader
	request := handshakeRequest{
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

	payload, err := rw.ReadBytes(0)
	if err != nil {
		return 0, err
	}
	payload = payload[:len(payload)-1]

	challenge, err := strconv.ParseInt(string(payload), 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(challenge), nil
}
