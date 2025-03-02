package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("read config: %s", err)
	}

	err = app.Run(&cfg)
	if err != nil {
		log.Fatalf("run app: %s", err)
	}
}
