package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	env     = "development"
	version = "1.0.0"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", healthCheckHandler)
	mux.HandleFunc("/", home)
	log.Println("start server at http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", env)
	fmt.Fprintf(w, "version: %s\n", version)
}
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello,home page"))
}
