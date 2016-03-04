package main

import (
	"net/http"
)

func handleCommand(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("This works!\n"))
}

func main() {

	http.HandleFunc("/", handleCommand)

	http.ListenAndServe(":80", nil)

}
