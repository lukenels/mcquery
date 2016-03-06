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
	SlackToken      string
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
	Hidden bool   // Whether every use can see output or not
	Type   string // "basic", "full" or "players"
}

func parseCommand(cmd string, cfg *Configuration) (*Command, error) {
	flagSet := flag.NewFlagSet("command", flag.ContinueOnError)
	command := new(Command)

	flagSet.BoolVar(&command.Hidden, "hidden", cfg.HiddenByDefault,
		"Determines if command output visible to everyone")

	flagSet.StringVar(&command.Ip, "ip", cfg.DefaultIp, "Minecraft server IP")

	flagSet.StringVar(&command.Type, "type", "basic", "Type of operation to do")

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

	switch command.Type {
	case "basic":
	case "full":
	case "players":
	default:
		return nil, errors.New(fmt.Sprintf("Unrecognized type \"%s\"", command.Type))
	}

	return command, nil
}

func handleCommand(input string, cfg *Configuration) (*Command, string, error) {

	cmd, err := parseCommand(input, cfg)
	if err != nil {
		return nil, "", err
	}

	rw, err, kill := mcquery.Connect(cmd.Ip, cmd.Port)
	if err != nil {
		kill <- true
		return cmd, "", err
	}
	challenge, err := mcquery.Handshake(rw)
	if err != nil {
		kill <- true
		return cmd, "", err
	}

	var responseString string

	switch cmd.Type {
	case "basic":
		statResponse, err := mcquery.BasicStat(rw, challenge)
		if err != nil {
			kill <- true
			return cmd, "", err
		}
		kill <- true

		responseString += fmt.Sprintf("```MOTD: %s\n", statResponse.Motd)
		responseString += fmt.Sprintf("Gametype: %s\n", statResponse.Gametype)
		responseString += fmt.Sprintf("Map: %s\n", statResponse.Map)
		responseString += fmt.Sprintf("NumPlayers: %s\n", statResponse.NumPlayers)
		responseString += fmt.Sprintf("MaxPlayers: %s\n", statResponse.MaxPlayers)
		responseString += fmt.Sprintf("HostPort: %d\n", statResponse.HostPort)
		responseString += fmt.Sprintf("HostIp: %s\n```", statResponse.HostIp)
	case "full":
		statResponse, err := mcquery.FullStat(rw, challenge)
		if err != nil {
			kill <- true
			return cmd, "", err
		}
		kill <- true

		responseString += "```\n"
		for k, v := range statResponse.KeyValues {
			responseString += fmt.Sprintf("%s: %s\n", k, v)
		}
		if len(statResponse.Players) > 0 {
			responseString += "Players:\n"
			for _, player := range statResponse.Players {
				responseString += fmt.Sprintf("    %s\n", player)
			}
		} else {
			responseString += "Players: <none>\n"
		}
		responseString += "```"
	case "players":
		statResponse, err := mcquery.FullStat(rw, challenge)
		if err != nil {
			kill <- true
			return cmd, "", err
		}
		kill <- true
		responseString += "```\n"
		if len(statResponse.Players) > 0 {
			for _, player := range statResponse.Players {
				responseString += fmt.Sprintf("%s\n", player)
			}
		} else {
			responseString += "<No Players Online>\n"
		}
		responseString += "```"
	default:
		return cmd, "", errors.New("Internal Error 42. Please Report this.")
	}

	return cmd, responseString, nil
}

func processRequest(w http.ResponseWriter, r *http.Request) {

	if globalConfiguration.SlackToken != "" &&
		globalConfiguration.SlackToken != r.FormValue("token") {
		log.Printf("Connection with bad token: %s and command: %s\n",
			r.FormValue("token"), r.FormValue("text"))
		w.WriteHeader(401) // unauthorized
		w.Write([]byte("Invalid API token. Your team is not set up to use this server"))
		return
	}

	cmd, responseString, err := handleCommand(r.FormValue("text"), &globalConfiguration)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	responseMap := make(map[string]interface{})
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
	debug := flag.String("debug", "", "Command to process for debug purposes")
	flag.Parse()

	log.Println("Loading configuration")
	globalConfiguration = getConfiguration(*configFile)
	log.Printf("Configuration is %+v\n", globalConfiguration)

	if *debug != "" {
		log.Printf("Running debug command \"%s\"\n", *debug)
		cmd, resp, err := handleCommand(*debug, &globalConfiguration)
		log.Printf("Command is %+v\n", cmd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", resp)
		return
	}

	http.HandleFunc("/", processRequest)

	log.Printf("Binding to port %s\n", *port)

	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)

	if err != nil {
		panic(err)
	}

}
