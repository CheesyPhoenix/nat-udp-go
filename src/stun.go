package src

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
)

type IPAndPort struct {
	IP   net.IP
	Port int
}

func GetIPAndPort() (*IPAndPort, error) {
	conn, err := net.Dial("udp4", "stun.l.google.com:19302")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := make([]byte, 20)
	binary.BigEndian.PutUint16(req[0:2], 0x0001)     // Message type = binding
	binary.BigEndian.PutUint16(req[2:4], 0)          // Message length
	binary.BigEndian.PutUint32(req[4:8], 0x21122442) // Message length
	rand.Read(req[8:20])

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	offset := 20
	for offset < bytesRead {
		attrType := binary.BigEndian.Uint16(buffer[offset : offset+2])
		attrLen := binary.BigEndian.Uint16(buffer[offset+2 : offset+4])
		attrValue := buffer[offset+4 : offset+4+int(attrLen)]

		if attrType != 0x0001 {
			offset += int(attrLen) + 4
			continue
		}

		return &IPAndPort{
			Port: int(binary.BigEndian.Uint16(attrValue[2:4])),
			IP:   net.IPv4(attrValue[4], attrValue[5], attrValue[6], attrValue[7]),
		}, nil
	}

	return nil, errors.New("unable to parse port and IP from STUN response")
}
