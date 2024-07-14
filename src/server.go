package src

import (
	"fmt"
	"net"
	"sync"
)

const ServerUDPPort = 12345
const ServerTCPPort = 4173

func Server() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: ServerUDPPort,
		Zone: "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, err = conn.WriteToUDP([]byte("UPD hole punch"), &net.UDPAddr{
		IP:   net.IPv4(89, 10, 217, 140),
		Port: ClientUDPPort,
		Zone: "",
	})
	if err != nil {
		fmt.Println("Hole punch err:", err.Error())
	}

	fmt.Printf("Listening on 0.0.0.0:%v\n", ServerUDPPort)
	fmt.Printf("Forwarding 127.0.0.1:%v\n", ServerTCPPort)

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
				Port: ServerTCPPort,
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
