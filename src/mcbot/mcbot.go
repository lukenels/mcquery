package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"mcquery"
	"net/http"
	"strconv"
	"strings"
)

func parseCommand(cmd string) (string, uint16, error) {
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return "", 0, errors.New("Your request was malformed.")
	} else if len(args) == 1 {
		return args[0], 25565, nil
	} else {
		port, err := strconv.ParseUint(args[1], 10, 16)
		if err != nil {
			return "", 0, errors.New("Your request was malformed.")
		}
		return args[0], uint16(port), nil
	}
}

func handleCommand(w http.ResponseWriter, r *http.Request) {

	ip, port, err := parseCommand(r.FormValue("text"))

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO consolidate all these steps in package mcquery
	rw, err, kill := mcquery.Connect(ip, port)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte("An error occurred while connecting to the MC server."))
		kill <- true
		return
	}
	challenge, err := mcquery.Handshake(rw)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte("An error occurred while handshaking with MC server."))
		kill <- true
		return
	}
	statResponse, err := mcquery.BasicStat(rw, challenge)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte("An error occurred while reading status of MC server."))
		kill <- true
		return
	}

	responseMap := make(map[string]interface{})
	responseString := fmt.Sprintf("```MOTD: %s\n", statResponse.Motd)
	responseString += fmt.Sprintf("Gametype: %s\n", statResponse.Gametype)
	responseString += fmt.Sprintf("Map: %s\n", statResponse.Map)
	responseString += fmt.Sprintf("NumPlayers: %s\n", statResponse.NumPlayers)
	responseString += fmt.Sprintf("MaxPlayers: %s\n", statResponse.MaxPlayers)
	responseString += fmt.Sprintf("HostPort: %d\n", statResponse.HostPort)
	responseString += fmt.Sprintf("HostIp: %s\n```", statResponse.HostIp)
	responseMap["text"] = responseString
	responseMap["response_type"] = "in_channel"
	data, err := json.Marshal(responseMap)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte("Internal JSON Error Ocurred."))
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(data)

}

func main() {

	port := flag.String("port", "80", "Port to bind to")
	flag.Parse()

	http.HandleFunc("/", handleCommand)

	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)

	if err != nil {
		panic(err)
	}

}
