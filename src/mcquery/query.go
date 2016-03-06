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

// Max amount of times looped readers can loop for, in case key values
// or player list is malformed
const loopLimit = 512

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

type fullStatRequest struct {
	Magic          [2]byte
	Type           byte
	SessionId      int32
	ChallengeToken int32
	Padding        [4]byte
}

type FullStatResponse struct {
	Type      byte
	SessionId int32
	Padding   [11]byte
	KeyValues map[string]string
	Padding2  [10]byte
	Players   []string
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

func Connect(ip string, port uint16) (*bufio.ReadWriter, error, chan<- bool) {
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

func FullStat(rw *bufio.ReadWriter, challenge int32) (*FullStatResponse, error) {

	var response FullStatResponse
	done := make(chan error)
	request := fullStatRequest{
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

		var header responseHeader
		err = binary.Read(rw, binary.BigEndian, &header)
		if err != nil {
			done <- err
			return
		}
		response.Type = header.Type
		response.SessionId = header.SessionId

		_, err = rw.Discard(11)
		if err != nil {
			done <- err
			return
		}

		response.KeyValues = make(map[string]string)
		for i := 0; ; i += 1 {
			var key, value []byte
			key, err = rw.ReadBytes(0)
			if err != nil {
				done <- err
				return
			}
			key = key[:len(key)-1]
			if len(key) == 0 {
				break
			}
			value, err = rw.ReadBytes(0)
			if err != nil {
				done <- err
				return
			}
			value = value[:len(value)-1]
			response.KeyValues[string(key)] = string(value)
			if i > loopLimit {
				done <- errors.New("Too many key values!")
				return
			}
		}

		_, err = rw.Discard(10)
		if err != nil {
			done <- err
			return
		}

		response.Players = make([]string, 0)
		for i := 0; ; i += 1 {

			playerName, err := rw.ReadBytes(0)
			if err != nil {
				done <- err
				return
			}
			playerName = playerName[:len(playerName)-1]
			if len(playerName) == 0 {
				break
			}
			response.Players = append(response.Players, string(playerName))

			if i > loopLimit {
				done <- errors.New("Too many players!")
				return
			}
		}

		done <- nil
	}()

	select {
	case <-time.After(2 * time.Second):
		return nil, nil
	case err := <-done:
		if err == nil {
			return &response, nil
		} else {
			return nil, err
		}
	}
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
		response.Type = responseHeader.Type
		response.SessionId = responseHeader.SessionId

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
	var challenge int32
	done := make(chan error)
	request := handshakeRequest{
		Magic:     magicBytes(),
		Type:      handshakeType,
		SessionId: 1, // TODO figure this out
	}

	go func() {
		err := binary.Write(rw, binary.BigEndian, request)
		if err != nil {
			done <- err
			return
		}
		if err = rw.Flush(); err != nil {
			done <- err
			return
		}
		if err = binary.Read(rw, binary.BigEndian, &response); err != nil {
			done <- err
			return
		}

		payload, err := rw.ReadBytes(0)
		if err != nil {
			done <- err
			return
		}
		payload = payload[:len(payload)-1]

		number, err := strconv.ParseInt(string(payload), 10, 32)
		if err != nil {
			done <- err
			return
		}
		challenge = int32(number)
		done <- nil
	}()

	select {
	case <-time.After(2 * time.Second):
		return 0, errors.New("Timed out")
	case err := <-done:
		if err == nil {
			return challenge, nil
		} else {
			return 0, err
		}
	}

}
