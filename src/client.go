package src

import (
	"fmt"
	"net"
)

func Client() {
	tcpServerConn, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 12346,
		Zone: "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tcpServerConn.Close()

	fmt.Println("Listening on 0.0.0.0:12346")

	for {
		conn, err := tcpServerConn.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func() {
			fmt.Println("New connection")
			defer fmt.Println("Connection closed")
			defer conn.Close()

			udpClientConn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
				IP:   net.IPv4(0, 0, 0, 0),
				Port: 12345,
				Zone: "",
			})
			if err != nil {
				fmt.Println(err)
				return
			}
			defer udpClientConn.Close()

			for {
				// From client
				buffer := make([]byte, 1024)
				bytesRead, err := conn.Read(buffer)

				if err != nil {
					fmt.Println("Got error reading tcp data: ", err)
					if err.Error() == "EOF" {
						break
					}
				}
				fmt.Println("Read ", bytesRead, " bytes from tcp client")
				udpClientConn.Write(buffer[0:bytesRead])
				if bytesRead > 0 {
					//fmt.Println("From tcp client: ", string(buffer))
				} else {
					continue
				}

				// From server
				for {
					buffer = make([]byte, 1024)
					bytesRead, err = udpClientConn.Read(buffer)
					if err != nil {
						fmt.Println("Got error reading upd data: ", err)
						if err.Error() == "EOF" {
							break
						}
					}
					fmt.Println("Read ", bytesRead, " bytes from udp server")
					conn.Write(buffer[0:bytesRead])
					if bytesRead == 0 {
						break
					}
				}
			}
		}()
	}
}
