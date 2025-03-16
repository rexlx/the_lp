package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	flag.Parse()
	db, err := NewPostgresDB(*dbLocation)
	if err != nil {
		log.Fatal(err)
	}
	app := NewApplication("localhost", db)
	sb := SoundBlockIn880Hz(time.Second)
	sb.PlaySound()
	log.Fatal(http.ListenAndServe(":8081", app.Gateway))
}
