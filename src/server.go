package src

import (
	"net"
	"sync"
	"time"
)

const ServerUDPPort = 12345
const ServerTCPPort = 4173

func StartHolePunch(conn net.UDPConn, addr net.UDPAddr, logLn func(string, ...any)) {
	go func() {
		for {
			_, err := conn.WriteToUDP([]byte(KeepAliveMessage), &addr)
			if err != nil {
				logLn("Hole punch err: %v", err.Error())
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

func Server(conn net.UDPConn, logLn func(string, ...any)) {
	logLn("Listening on 0.0.0.0:%v", ServerUDPPort)
	logLn("Forwarding 127.0.0.1:%v", ServerTCPPort)

	//tcpConnections := make(map[net.Addr]*net.TCPConn)
	tcpConnections := new(sync.Map)

	for {
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			logLn("Got error reading from connection: %v", err.Error())
		}
		if bytesRead == KeepAliveMessageLength && string(buffer[0:bytesRead]) == KeepAliveMessage {
			logLn("Received Keep-Alive message")
			continue
		}
		logLn("Read %v bytes from udp client", bytesRead)

		val, ok := tcpConnections.Load(addr.String())
		var tcpClientConn *net.TCPConn
		if !ok {
			tcpClientConn, err = net.DialTCP("tcp4", nil, &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: ServerTCPPort,
				Zone: "",
			})
			if err != nil {
				logLn("%v", err.Error())
				return
			}

			go func() {
				for {
					buffer := make([]byte, 1024)
					bytesRead, err := tcpClientConn.Read(buffer)
					logLn("Read %v bytes from tcp server", bytesRead)
					conn.WriteTo(buffer[0:bytesRead], addr)

					if err != nil {
						logLn("Got error reading from connection: %v", err.Error())
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
			logLn("From upd client: %v", string(buffer[0:bytesRead]))
		}
	}
}
