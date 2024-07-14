package main

import (
	"fmt"
	"net"
	"os"

	"github.com/cheesyphoenix/nat-udp-go/src"
)

func main() {
	if os.Args[1] == "server" {
		ipAndPort, err := src.GetIPAndPort(&net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: src.ServerUDPPort,
			Zone: "",
		})
		if err != nil {
			fmt.Println("Unable to get IP and port:", err)
			return
		}
		fmt.Println("---------------------------")
		fmt.Println("IP:", ipAndPort.IP)
		fmt.Println("Port:", ipAndPort.Port)
		fmt.Println("---------------------------")

		src.Server()
	} else if os.Args[1] == "client" {
		ipAndPort, err := src.GetIPAndPort(&net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: src.ClientUDPPort,
			Zone: "",
		})
		if err != nil {
			fmt.Println("Unable to get IP and port:", err)
			return
		}
		fmt.Println("---------------------------")
		fmt.Println("IP:", ipAndPort.IP)
		fmt.Println("Port:", ipAndPort.Port)
		fmt.Println("---------------------------")

		src.Client()
	}
}
