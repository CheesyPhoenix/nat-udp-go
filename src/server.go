package src

import (
	"net"
	"sync"
	"time"
)

const ServerUDPPort = 12347
const ServerTCPPort = 5173

func StartHolePunch(conn *ReliableUDPConn, addr net.UDPAddr, logLn func(string, ...any)) {
	go func() {
		for {
			_, err := conn.WriteKeepAliveTo(&addr)
			if err != nil {
				logLn("Hole punch err: %v", err.Error())
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

func Server(conn *ReliableUDPConn, logLn func(string, ...any)) {
	logLn("Listening on 0.0.0.0:%v", ServerUDPPort)
	logLn("Forwarding 127.0.0.1:%v", ServerTCPPort)

	//tcpConnections := make(map[net.Addr]*net.TCPConn)
	tcpConnections := new(sync.Map)

	for {
		packets, addr, err := conn.ReadFrom(logLn)
		if err != nil {
			logLn("Got error reading from connection: %v", err.Error())
		}

		logLn("Read %v packets from udp client", len(packets))

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

		for _, packet := range packets {
			tcpClientConn.Write(packet)
		}
	}
}
