package main

import (
	"github.com/ilyakaznacheev/gochan"
)

func main() {
	// create server with default config
	s := gochan.NewServer(nil)

	// and just run it
	s.Run()
}
