package src

import (
	"fmt"
	"net"
)

const ClientUDPPort = 12344
const ClientTCPPort = 12346

func Client(serverAddr net.UDPAddr) {
	tcpServerConn, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: ClientTCPPort,
		Zone: "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tcpServerConn.Close()

	udpClientConn, err := net.DialUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: ClientUDPPort,
		Zone: "",
	}, &serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer udpClientConn.Close()

	fmt.Printf("Listening on 0.0.0.0:%v\n", ClientTCPPort)

	_, err = udpClientConn.Write([]byte("UPD hole punch"))
	if err != nil {
		fmt.Println("Hole punch err:", err.Error())
	}

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
		fmt.Println("Handling connection")

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
