package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/DHBosworth/technichalexercise/backend"
	"github.com/DHBosworth/technichalexercise/service"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	flag.Parse()
}

var addr = flag.String("p", "8080", "Port the server will listen on")

func main() {
	data, err := backend.NewMongoDataSource("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Unable to create mongo db data source: %v", err)
	}
	log.Debugf("Connected to data source")

	log.Debugf("Starting Server on port :%s", *addr)
	microService := service.New(data, nil)
	err = http.ListenAndServe(":"+*addr, microService)
	if err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
