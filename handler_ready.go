package main

import (
	"io"
	"net/http"
)

func handleReady(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	// Body
	io.WriteString(response, "OK")
}