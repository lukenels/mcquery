package main

import (
	"bufio"
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

	buffer := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	challenge, err := mcquery.Handshake(buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("The server responded with challenge id of %d\n", challenge)

	resp, err := mcquery.BasicStat(buffer, challenge)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Server MOTD: %s\n", resp.Motd)
	fmt.Printf("Current Players: %s\n", resp.NumPlayers)
	fmt.Printf("Server Max Players: %s\n", resp.MaxPlayers)
	fmt.Printf("Server Game Type: %s\n", resp.Gametype)
	fmt.Printf("Server Map Name: %s\n", resp.Map)
	fmt.Printf("Server IP: %s\n", resp.HostIp)
	fmt.Printf("Server Port: %d\n", resp.HostPort)

	conn.Close()
}
