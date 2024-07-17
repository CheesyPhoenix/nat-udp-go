package main

import (
	"fmt"
	"net"
	"os"

	"github.com/cheesyphoenix/nat-udp-go/src"
	"github.com/cheesyphoenix/nat-udp-go/src/tui"
)

// TODO: Proper interface + stop sending hole-punch requests when connected

func main() {
	if len(os.Args) < 2 {
		tui.StartTUI()
	} else if os.Args[1] == "server" {
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
		fmt.Printf("Your address: %v:%v\n", ipAndPort.IP, ipAndPort.Port)
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
		fmt.Printf("Your address: %v:%v\n", ipAndPort.IP, ipAndPort.Port)
		fmt.Println("---------------------------")

		serverAddr, err := tui.PromptForAddress()
		if err != nil {
			fmt.Println(err)
			return
		}

		src.Client(*serverAddr)
	}
}
