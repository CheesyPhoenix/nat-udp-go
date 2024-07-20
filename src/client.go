package src

import (
	"net"
	"time"
)

const ClientUDPPort = 12345
const ClientTCPPort = 12346

const KeepAliveMessage = "UDP Hole Punch"
const KeepAliveMessageLength = len(KeepAliveMessage)

func Client(serverAddr net.UDPAddr, logLn func(string, ...any)) {
	tcpServerConn, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: ClientTCPPort,
		Zone: "",
	})
	if err != nil {
		logLn("%v", err.Error())
		return
	}
	defer tcpServerConn.Close()

	udpClientConn, err := net.DialUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: ClientUDPPort,
		Zone: "",
	}, &serverAddr)
	if err != nil {
		logLn("%v", err.Error())
		return
	}
	defer udpClientConn.Close()

	logLn("Listening on 0.0.0.0:%v", ClientTCPPort)

	go func() {
		for {
			_, err := udpClientConn.Write([]byte(KeepAliveMessage))
			if err != nil {
				logLn("Hole punch err: %v", err.Error())
			}

			time.Sleep(time.Second * 5)
		}
	}()

	tcpConnections := make(chan net.Conn, 100)

	go func() {
		for {
			conn, err := tcpServerConn.Accept()
			if err != nil {
				logLn("%v", err.Error())
				continue
			}
			tcpConnections <- conn
			logLn("New connection")
		}
	}()

	for {
		conn := <-tcpConnections
		logLn("Handling connection")

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
					logLn("Got error reading tcp data: %v", err.Error())
					if err.Error() == "EOF" {
						break
					}
				}
				logLn("Read %v bytes from tcp client", bytesRead)
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
					logLn("Got error reading upd data: %v", err.Error())
					if err.Error() == "EOF" {
						break
					}
				}
				if bytesRead == KeepAliveMessageLength && string(buffer[0:bytesRead]) == KeepAliveMessage {
					logLn("Received keep-alive message")
					continue
				}

				logLn("Read %v bytes from udp server", bytesRead)
				if bytesRead == 0 {
					break
				}
				conn.Write(buffer[0:bytesRead])
			}

			stop <- true
		}()

		<-stop
		stop <- true

		conn.Write([]byte{})
		conn.Close()
		logLn("Connection closed")
	}
}
