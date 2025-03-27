package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", handleReady)

	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}