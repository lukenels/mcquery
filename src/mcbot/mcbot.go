package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"mcquery"
	"net/http"
	"os"
	"strings"
)

type Configuration struct {
	HiddenByDefault bool
	DefaultPort     uint16
	DefaultIp       string
}

var globalConfiguration Configuration

func getConfiguration(filename string) Configuration {
	var cfg Configuration
	if filename == "" {
		log.Println("No configuration file, using default")
		return Configuration{
			false,
			25565, // default minecraft port
			"",
		}
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err.Error())
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cfg
}

type Command struct {
	Ip     string
	Port   uint16
	Hidden bool // Whether every use can see output or not
}

func parseCommand(cmd string, cfg Configuration) (*Command, error) {
	flagSet := flag.NewFlagSet("command", flag.ContinueOnError)
	command := new(Command)

	flagSet.BoolVar(&command.Hidden, "hidden", cfg.HiddenByDefault,
		"Determines if command output visible to everyone")

	flagSet.StringVar(&command.Ip, "ip", cfg.DefaultIp, "Minecraft server IP")

	var portNum uint64
	flagSet.Uint64Var(&portNum, "port", uint64(cfg.DefaultPort), "MC Server Port")

	err := flagSet.Parse(strings.Fields(cmd))
	if err != nil {
		return nil, err
	}
	if portNum > math.MaxUint16 {
		return nil, errors.New("Port number out of range")
	}

	command.Port = uint16(portNum)

	return command, nil
}

func handleCommand(w http.ResponseWriter, r *http.Request) {

	cmd, err := parseCommand(r.FormValue("text"), globalConfiguration)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO consolidate all these steps in package mcquery
	rw, err, kill := mcquery.Connect(cmd.Ip, cmd.Port)
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
	kill <- true

	responseMap := make(map[string]interface{})
	responseString := fmt.Sprintf("```MOTD: %s\n", statResponse.Motd)
	responseString += fmt.Sprintf("Gametype: %s\n", statResponse.Gametype)
	responseString += fmt.Sprintf("Map: %s\n", statResponse.Map)
	responseString += fmt.Sprintf("NumPlayers: %s\n", statResponse.NumPlayers)
	responseString += fmt.Sprintf("MaxPlayers: %s\n", statResponse.MaxPlayers)
	responseString += fmt.Sprintf("HostPort: %d\n", statResponse.HostPort)
	responseString += fmt.Sprintf("HostIp: %s\n```", statResponse.HostIp)
	responseMap["text"] = responseString
	if cmd.Hidden {
		responseMap["response_type"] = "ephemeral"
	} else {
		responseMap["response_type"] = "in_channel"
	}
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

	log.Println("Starting mcbot server")

	port := flag.String("port", "80", "Port to bind to")
	configFile := flag.String("config", "", "Configuration file")
	flag.Parse()

	log.Println("Loading configuration")
	globalConfiguration = getConfiguration(*configFile)
	log.Printf("Configuration is %+v\n", globalConfiguration)

	http.HandleFunc("/", handleCommand)

	log.Printf("Binding to port %s\n", *port)

	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)

	if err != nil {
		panic(err)
	}

}
