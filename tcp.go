package main

import "encoding/binary"

/*
	 ############################
			TCP SECTION
	   ############################
*/
const (
	SYN     byte = 0b00000001
	ACK     byte = 0b00000010
	SYN_ACK byte = 0b00000100
)

type TCPPacket struct {
	flags           byte
	sequence        uint32
	acknowledgement uint32
	data            []byte
}

func create_tcp_packet(flags byte, sequence uint32, acknowledgement uint32, data []byte) *TCPPacket {
	return &TCPPacket{
		flags:           flags,
		sequence:        sequence,
		acknowledgement: acknowledgement,
		data:            data,
	}
}

func unpack(buf []byte) *TCPPacket {
	packet := &TCPPacket{}
	packet.flags = buf[0]

	packet.sequence = binary.BigEndian.Uint32(buf[1:5])
	packet.acknowledgement = binary.BigEndian.Uint32(buf[5:9])
	packet.data = buf[9:]
	return packet
}

func pack(toPack *TCPPacket) []byte {
	packet := make([]byte, 512)
	packet[0] = toPack.flags
	sequenceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceBytes, toPack.sequence)
	copy(packet[1:], sequenceBytes)

	acknowledgementBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(acknowledgementBytes, toPack.acknowledgement)
	copy(packet[5:], acknowledgementBytes)

	copy(packet[9:], toPack.data)

	return packet
}
