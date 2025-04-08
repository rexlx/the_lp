package main

import (
	"flag"
	"log"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	flag.Parse()
	db, err := NewPostgresDB(*dbLocation)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	app := NewApplication("http://localhost:8081", db)
	app.Logger = logger
	// sb := SoundBlockIn880Hz(time.Second)
	// sb.PlaySound()
	log.Fatal(http.ListenAndServe(":8081", app.Gateway))
}
