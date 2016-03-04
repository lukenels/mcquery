package main

import (
	"bufio"
	//	"encoding/binary"
	"flag"
	"fmt"
	"mcquery"
	"net"
	"os"
)

func main() {
	ipString := flag.String("ip", "127.0.0.1", "IP of Server")
	port := flag.Uint("port", 25565, "Port of Server")
	flag.Parse()

	if _, err := net.ResolveIPAddr("ip", *ipString); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	fmt.Printf("Using %s as ip and %d as port\n", *ipString, *port)
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *ipString, *port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	// query := mcquery.SendHeader{
	// 	Magic:     [2]byte{0xFE, 0xFD},
	// 	Type:      0x09,
	// 	SessionId: 1,
	// }

	buffer := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	mcquery.Handshake(buffer)

	// binary.Write(buffer, binary.BigEndian, query)
	// buffer.Flush()
	//
	// var response mcquery.RecvHeader
	//
	// binary.Read(buffer, binary.BigEndian, &response)
	//
	// s, _ := buffer.ReadString(0)
	//
	// fmt.Println(s)

	conn.Close()
}
