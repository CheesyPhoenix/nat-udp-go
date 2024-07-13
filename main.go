package main

import (
	"fmt"
	"os"

	"github.com/cheesyphoenix/nat-udp-go/src"
)

func main() {
	ipAndPort, err := src.GetIPAndPort()
	if err != nil {
		fmt.Println("Unable to get IP and port:", err)
		return
	}
	fmt.Println("---------------------------")
	fmt.Println("IP:", ipAndPort.IP)
	fmt.Println("Port:", ipAndPort.Port)
	fmt.Println("---------------------------")

	if os.Args[1] == "server" {
		src.Server()
	} else if os.Args[1] == "client" {
		src.Client()
	}
}
