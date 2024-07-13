package src

import (
	"fmt"
	"net"
)

func Server() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 12345,
		Zone: "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Listening on 0.0.0.0:12345")

	for {
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Got error reading from connection: ", err)
		}
		fmt.Println("Read ", bytesRead, " bytes from udp client")

		go func() {
			tcpClientConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 25565,
				Zone: "",
			})
			if err != nil {
				fmt.Println(err)
				return
			}
			defer tcpClientConn.Close()

			tcpClientConn.Write(buffer[0:bytesRead])
			if bytesRead > 0 {
				fmt.Println("From upd client: ", string(buffer))
			}

			for {
				buffer = make([]byte, 1024)
				bytesRead, err = tcpClientConn.Read(buffer)
				if err != nil {
					fmt.Println("Got error reading from connection: ", err)
				}
				fmt.Println("Read ", bytesRead, " bytes from tcp server")
				conn.WriteTo(buffer[0:bytesRead], addr)
				if bytesRead == 0 {
					break
				}
			}
		}()
	}
}
