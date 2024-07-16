package tui

import (
	"fmt"
	"net"
)

func PromptForAddress() (*net.UDPAddr, error) {
	fmt.Print("Enter remote address: ")

	var addrStr string
	_, err := fmt.Scanln(&addrStr)
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("got error trying to read line: %v", err.Error())
	}

	addr, err := net.ResolveUDPAddr("udp4", addrStr)
	if err != nil {
		fmt.Println("Address is not valid. Please try again. Error:", err)
		return PromptForAddress()
	}
	return addr, nil
}
