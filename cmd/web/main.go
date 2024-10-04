package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/delete", snippetCreate)

	log.Println("Starting server on port :3000")
	err := http.ListenAndServe(":3000", mux)
	log.Fatal(err) // NOTE: ListenAndServe always return a not-nil error so no need to put a if err != nil check here
}
