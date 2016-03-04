package main

import (
	"flag"
	"fmt"
	"mcquery"
)

func main() {
	ipString := flag.String("ip", "127.0.0.1", "IP of Server")
	port := flag.Uint("port", 25565, "Port of Server")
	flag.Parse()

	buffer, err, killConn := mcquery.Connect(*ipString, uint16(*port))
	if err != nil {
		panic(err)
	}

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

	killConn <- true // Stupid because main is about to exit

}
