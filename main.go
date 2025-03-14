package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	flag.Parse()
	db, err := NewPostgresDB(*dbLocation)
	if err != nil {
		log.Fatal(err)
	}
	app := NewApplication("localhost", db)
	log.Fatal(http.ListenAndServe(":8080", app.Gateway))
}
