package src

import (
	"bytes"
	"net"
	"reflect"
	"testing"
)

func TestHeaderToBytes(t *testing.T) {
	header := Header{
		Version:  1,
		PacketId: 0,
		Flags: Flags{
			IsKeepAlive: true,
		},
	}

	header_bytes := header.ToBytes()
	expected_bytes := []byte{0x01, 0b10000000, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(header_bytes, expected_bytes) {
		t.Fatalf("Bytes are not equal. %v != %v", header_bytes, expected_bytes)
	}
}

func TestHeaderFromBytes(t *testing.T) {
	header, err := ParseHeader([]byte{0x01, 0b10000000, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		t.Fatalf("Received error trying to parse a valid header: %v", err)
	}

	expected_header := Header{
		Version:  1,
		PacketId: 0,
		Flags: Flags{
			IsKeepAlive: true,
		},
	}

	if !reflect.DeepEqual(*header, expected_header) {
		t.Fatalf("Headers are not equal. %v != %v", header, expected_header)
	}
}

func TestWriteKeepAlive(t *testing.T) {
	serverConn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 55550})
	if err != nil {
		t.Fatalf("Received error trying to start test udp server: %v", err)
	}

	clientConn, err := net.DialUDP(
		"udp4",
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)},
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 55550},
	)
	if err != nil {
		t.Fatalf("Received error trying to start test udp client: %v", err)
	}

	reliableClient := NewReliableUDPConn(*clientConn)
	_, err = reliableClient.WriteKeepAlive()
	if err != nil {
		t.Fatalf("Received error trying to write: %v", err)
	}

	buffer := make([]byte, 1024)
	bytesRead, err := serverConn.Read(buffer)
	if err != nil {
		t.Fatalf("Received error trying to read: %v", err)
	}
	if bytesRead != HeaderSize {
		t.Fatalf("Read incorrect number of bytes. Read: %v, expected: %v", bytesRead, HeaderSize)
	}

	header, err := ParseHeader(buffer)
	if err != nil {
		t.Fatalf("Received error trying to parse header: %v", err)
	}
	if !header.Flags.IsKeepAlive {
		t.Fatalf("IsKeepAlive is not set in header: %v", header)
	}
}
