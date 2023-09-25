package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	SYN     byte = 0b00000001
	ACK     byte = 0b00000010
	FIN     byte = 0b00000100
	DATA    byte = 0b00001000
	SYN_ACK byte = 0b00000011
)

func main() {
	go server(25555)
	go server(25554)
	time.Sleep(4 * time.Second)
	go client(25555)
	go client(25554)
	time.Sleep(50 * time.Second)	
}

/* ############################
		TCP SECTION
   ############################ */
   
type TCPPacket struct {
	flags 		 	byte
	sequence      	uint32
	acknowledgement uint32
	data          	[]byte
}

func create_tcp_packet(flags byte, sequence uint32, acknowledgement uint32, data []byte) *TCPPacket {
	return &TCPPacket{
		flags:          flags,
		sequence:       sequence,
		acknowledgement: acknowledgement,
		data:           data,
	}
}

func unpack(buf []byte) *TCPPacket {
	packet := &TCPPacket{}
	packet.flags = buf[0] 
	
	packet.sequence = binary.BigEndian.Uint32(buf[1:5])
	packet.acknowledgement = binary.BigEndian.Uint32(buf[5:9])
	packet.data = buf[9:]
	return packet;
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
	
	return packet;
}

/* ############################
		CLIENT SECTION
   ############################ */

func client(port int) {
	sequence := rand.Uint32()
	var sequence_ptr *uint32 = &sequence
	
	// Establish a connection
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer conn.Close()
	for {
		err := clientThreeWayHandshake(conn, sequence_ptr)
		if (err != nil) {
			fmt.Println("Error in handshake:", err)
		} else {
			break;
		}
	}
}

func clientThreeWayHandshake(conn *net.UDPConn, sequence *uint32) error {
	
	// Send SYN packet
	err := sendSyn(conn, *sequence)
	if (err != nil) {return err}
	*sequence++
	
	recv_packet, err := receiveAndVerifySynAck(conn, *sequence)
	if (err != nil) {return err}

	err = clientSendAck(recv_packet, conn, *sequence)
	if (err != nil) {return err}
	*sequence++
	
	fmt.Println("Handshake completed with server.")
	return nil
}

func sendSyn( conn *net.UDPConn, sequence uint32) error {	
	packet := &TCPPacket {
		flags: SYN,
		sequence: sequence,
	}
	
	_, err := conn.Write(pack(packet))
	if err != nil {return err}
	fmt.Println("Sending SYN to: ", conn.LocalAddr().String())
	return nil
}

func receiveAndVerifySynAck( conn *net.UDPConn, sequence uint32) (*TCPPacket, error) {
	buf := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil || n == 0 {
		return nil, err
	}
	recv_packet := unpack(buf)
	if recv_packet.flags != SYN_ACK {
		return nil, fmt.Errorf("Wrong packet received");
	}
	
	fmt.Println("SYN_ACK received from: ", conn.LocalAddr().String())
	if (recv_packet.acknowledgement != sequence) {
		return nil, fmt.Errorf("SYN_ACK ack number did not match sequence");
	}
	return recv_packet, nil
}

func clientSendAck(recv_packet *TCPPacket, conn *net.UDPConn, sequence uint32) error {
	resp_packet := &TCPPacket {
		flags: ACK,
		acknowledgement: recv_packet.sequence+1,
		sequence: sequence,
	}

	fmt.Println("sending ACK with acknowledgement")
	_, err := conn.Write(pack(resp_packet))
	if err != nil {return err}
	return nil
}

/* ############################
		SERVER SECTION
   ############################ */

func server(port int) {
	sequenceMap := make(map[string]uint32)
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Print(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Print(err)
	}

	for {
		var buf [512]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Print(err)
		}

		recv_packet := unpack(buf[0:])
		if (recv_packet.flags == SYN) {
			handleSyn(recv_packet, sequenceMap, addr, conn)
		} else if (recv_packet.flags == ACK) {
			handleAck(recv_packet, sequenceMap, addr, conn)
		}
	}
}

func handleSyn(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {
	
	fmt.Println("SYN received")
	sequenceMap[addr.String()] = rand.Uint32();
	resp_packet := &TCPPacket {
		flags: SYN_ACK,
		acknowledgement: recv_packet.sequence+1,
		sequence: sequenceMap[addr.String()],
	}
	_, err := conn.WriteToUDP(pack(resp_packet), addr)
	sequenceMap[addr.String()]++
	if err != nil {
		fmt.Println("Error sending SYN_ACK:", err)
		return
	}
}

func handleAck(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {
	if (recv_packet.acknowledgement != sequenceMap[addr.String()]) {
		fmt.Println("Ack does not match server sequence IP: " + addr.String())
	}
	fmt.Println("Connection established with IP: " + addr.String())
}
