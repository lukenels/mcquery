package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func handleCommand(w http.ResponseWriter, r *http.Request) {

	user := r.FormValue("user_name")

	responseMap := make(map[string]interface{})
	responseMap["text"] = fmt.Sprintf("Hi %s", user)
	responseMap["response_type"] = "in_channel"
	data, err := json.Marshal(responseMap)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(400)
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
