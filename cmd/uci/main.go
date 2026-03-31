package main

import (
	"chessV2/internal/engine"
	"chessV2/internal/uci"
	"log"
	"os"
)

func main() {
	srv, err := uci.NewServer(engine.NewEngine())
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
