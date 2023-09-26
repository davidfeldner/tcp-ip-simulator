package main

import "encoding/binary"

/*
	 ############################
			TCP SECTION
	   ############################
*/

const (
	MethodGET    byte = 0b00000001
	MethodPOST   byte = 0b00000010
	MethodUPDATE byte = 0b00000100
)

type HTTPPacket struct {
	reqType         byte
	sequence        uint32
	acknowledgement uint32
	data            []byte
}

func unpackHTTP(buf []byte) *HTTPPacket {
	packet := &HTTPPacket{}
	packet.reqType = buf[0]

	packet.sequence = binary.BigEndian.Uint32(buf[1:5])
	packet.acknowledgement = binary.BigEndian.Uint32(buf[5:9])
	packet.data = buf[9:]
	return packet
}

func packHTPP(toPack *HTTPPacket) []byte {
	packet := make([]byte, 512)
	packet[0] = toPack.reqType
	sequenceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceBytes, toPack.sequence)
	copy(packet[1:], sequenceBytes)

	acknowledgementBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(acknowledgementBytes, toPack.acknowledgement)
	copy(packet[5:], acknowledgementBytes)

	copy(packet[9:], toPack.data)

	return packet
}
