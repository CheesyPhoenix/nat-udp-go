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

	fmt.Println("Listening on 0.0.0.0:12346")

	tcpConnections := make(chan net.Conn, 100)

	go func() {
		for {
			conn, err := tcpServerConn.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}
			tcpConnections <- conn
			fmt.Println("New connection")
		}
	}()

	for {
		conn := <-tcpConnections
		fmt.Println("Handeling connection")

		stop := make(chan bool, 10)

		go func() {
			// From client
			for {
				select {
				case <-stop:
					stop <- true
					return
				default:
				}

				buffer := make([]byte, 1024)
				bytesRead, err := conn.Read(buffer)

				if err != nil {
					fmt.Println("Got error reading tcp data: ", err)
					if err.Error() == "EOF" {
						break
					}
				}
				fmt.Println("Read ", bytesRead, " bytes from tcp client")
				if bytesRead == 0 {
					break
				}
				udpClientConn.Write(buffer[0:bytesRead])
			}

			stop <- true
		}()

		go func() {
			// From server
			for {
				select {
				case <-stop:
					stop <- true
					return
				default:
				}

				buffer := make([]byte, 1024)
				bytesRead, err := udpClientConn.Read(buffer)
				if err != nil {
					fmt.Println("Got error reading upd data: ", err)
					if err.Error() == "EOF" {
						break
					}
				}
				fmt.Println("Read ", bytesRead, " bytes from udp server")
				if bytesRead == 0 {
					break
				}
				conn.Write(buffer[0:bytesRead])
			}

			stop <- true
		}()

		<-stop
		stop <- true

		conn.Close()
		fmt.Println("Connection closed")
	}
}
