package main

import (
	"log"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/bootstrap"
)

func main() {
	if err := bootstrap.NewApp().Run(); err != nil {
		log.Fatal(err)
	}
}
