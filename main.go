package main

import (
	"os"

	"github.com/cheesyphoenix/nat-udp-go/src"
)

func main() {
	if os.Args[1] == "server" {
		src.Server()
	} else if os.Args[1] == "client" {
		src.Client()
	}
}
