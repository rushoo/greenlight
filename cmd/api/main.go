package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API  port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	app := &application{
		config: cfg,
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}

	//自定义server以使用自定义的port
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.route(),
	}
	log.Printf("starting %s  on %s", cfg.env, srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
