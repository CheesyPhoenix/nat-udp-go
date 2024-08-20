package src

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Flags struct {
	IsKeepAlive   bool
	ReservedFlag2 bool
	ReservedFlag3 bool
	ReservedFlag4 bool
	ReservedFlag5 bool
	ReservedFlag6 bool
	ReservedFlag7 bool
	ReservedFlag8 bool
}

func (flags *Flags) ToByte() byte {
	var res byte = 0x00
	if flags.IsKeepAlive {
		res |= 0b10000000
	}

	return res
}
func NewFlagsFromByte(b byte) Flags {
	return Flags{
		IsKeepAlive: 0b10000000&b == 0b10000000,
	}
}

type Header struct {
	Version  uint8
	Flags    Flags
	PacketId uint32
}

const HeaderSize = 6
const CurrentVersion uint8 = 1

func (header *Header) ToBytes() []byte {
	res := []byte{byte(header.Version), header.Flags.ToByte()}
	binary.BigEndian.AppendUint32(res, header.PacketId)
	return res
}

func ParseHeader(bytes []byte) (*Header, error) {
	if len(bytes) < HeaderSize {
		return nil, fmt.Errorf("too few bytes supplied. A minimum of %v bytes are required", HeaderSize)
	}

	return &Header{
		Version:  uint8(bytes[0]),
		Flags:    NewFlagsFromByte(bytes[1]),
		PacketId: binary.BigEndian.Uint32(bytes[2:6]),
	}, nil
}

type ReliableUDPConn struct {
	UDPConn net.UDPConn
	//Address -> PacketId
	NextIncomingPacketIDs map[string]uint32
	NextOutgoingPacketIDs map[string]uint32
	//Address -> PacketId -> data
	FutureIncomingPackets map[string]map[uint32][]byte
}

func (conn *ReliableUDPConn) ReadFrom(logLn func(string, ...any)) (packets [][]byte, addr net.Addr, err error) {
	for {
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.UDPConn.ReadFrom(buffer)
		if err != nil {
			return nil, nil, err
		}
		if bytesRead < HeaderSize {
			// Discard the packet if it does not include a header
			logLn("Received malformed packet with only %v of the required %v bytes", bytesRead, HeaderSize)
			continue
		}

		header, err := ParseHeader(buffer)
		if err != nil {
			return nil, nil, err
		}
		if header.Version != CurrentVersion {
			logLn("Received packet with invalid version from %v. Current version is %v, but received version is %v", addr.String(), CurrentVersion, header.Version)
			continue
		}
		if header.Flags.IsKeepAlive {
			logLn("Received keep-alive packet from %v", addr.String())
			continue
		}

		data := buffer[HeaderSize:bytesRead]

		futurePackets := conn.FutureIncomingPackets[addr.String()]
		nextPacketId := conn.NextIncomingPacketIDs[addr.String()]

		if header.PacketId != nextPacketId {
			if futurePackets != nil {
				futurePackets[header.PacketId] = data
			} else {
				conn.FutureIncomingPackets[addr.String()] = map[uint32][]byte{header.PacketId: data}
			}
			continue
		}

		packets := [][]byte{data}

		if futurePackets != nil {
			var offset uint32 = 1
			for {
				packet := futurePackets[header.PacketId+offset]
				if packet != nil {
					delete(futurePackets, header.PacketId+offset)
					packets = append(packets, packet)
				} else {
					break
				}
			}
		}

		conn.NextIncomingPacketIDs[addr.String()] = header.PacketId + 1

		return packets, addr, nil
	}
}

func (conn *ReliableUDPConn) Read(logLn func(string, ...any)) (packets [][]byte, err error) {
	for {
		packets, addr, err := conn.ReadFrom(logLn)
		if err != nil {
			return nil, err
		}
		if (addr).String() != conn.UDPConn.RemoteAddr().String() {
			continue
		}
		return packets, nil
	}
}

func (conn *ReliableUDPConn) WriteTo(data []byte, addr net.Addr) (int, error) {
	packetId := conn.NextOutgoingPacketIDs[addr.String()]
	conn.NextOutgoingPacketIDs[addr.String()] += 1

	header := Header{
		Version:  1,
		PacketId: packetId,
	}

	return conn.UDPConn.WriteTo(append(header.ToBytes(), data...), addr)
}

func (conn *ReliableUDPConn) Write(data []byte) (int, error) {
	packetId := conn.NextOutgoingPacketIDs[conn.UDPConn.RemoteAddr().String()]
	conn.NextOutgoingPacketIDs[conn.UDPConn.RemoteAddr().String()] += 1

	header := Header{
		Version:  1,
		PacketId: packetId,
	}

	return conn.UDPConn.Write(append(header.ToBytes(), data...))
}

func (conn *ReliableUDPConn) WriteKeepAliveTo(addr net.Addr) (int, error) {
	header := Header{
		Version:  1,
		PacketId: 0,
		Flags: Flags{
			IsKeepAlive: true,
		},
	}

	return conn.UDPConn.WriteTo(header.ToBytes(), addr)
}

func (conn *ReliableUDPConn) WriteKeepAlive() (int, error) {
	header := Header{
		Version:  1,
		PacketId: 0,
		Flags: Flags{
			IsKeepAlive: true,
		},
	}

	return conn.UDPConn.Write(header.ToBytes())
}

func NewReliableUDPConn(conn net.UDPConn) ReliableUDPConn {
	return ReliableUDPConn{
		UDPConn:               conn,
		NextIncomingPacketIDs: make(map[string]uint32),
		NextOutgoingPacketIDs: make(map[string]uint32),
		FutureIncomingPackets: make(map[string]map[uint32][]byte),
	}
}
