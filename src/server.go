package src

import (
	"fmt"
	"net"
	"sync"
)

func Server() {
	holeConn, _ := net.DialUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 12345,
		Zone: "",
	}, &net.UDPAddr{
		IP:   net.IPv4(89, 10, 217, 140),
		Port: 54250,
		Zone: "",
	})
	holeConn.Close()

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

	//tcpConnections := make(map[net.Addr]*net.TCPConn)
	tcpConnections := new(sync.Map)

	for {
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Got error reading from connection: ", err)
		}
		fmt.Println("Read ", bytesRead, " bytes from udp client")

		val, ok := tcpConnections.Load(addr.String())
		var tcpClientConn *net.TCPConn
		if !ok {
			tcpClientConn, err = net.DialTCP("tcp4", nil, &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 25566,
				Zone: "",
			})
			if err != nil {
				fmt.Println(err)
				return
			}

			go func() {
				for {
					buffer = make([]byte, 1024)
					bytesRead, err = tcpClientConn.Read(buffer)
					fmt.Println("Read ", bytesRead, " bytes from tcp server")
					conn.WriteTo(buffer[0:bytesRead], addr)

					if err != nil {
						fmt.Println("Got error reading from connection: ", err)
						if err.Error() == "EOF" {
							tcpClientConn.Close()
							tcpConnections.Delete(addr.String())
							return
						}
					}
				}
			}()

			tcpConnections.Store(addr.String(), &tcpClientConn)
		} else {
			tcpClientConn = *val.(**net.TCPConn)
		}

		tcpClientConn.Write(buffer[0:bytesRead])
		if bytesRead > 0 {
			fmt.Println("From upd client: ", string(buffer))
		}
	}
}
