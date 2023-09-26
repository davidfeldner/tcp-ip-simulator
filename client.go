package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

/* ############################
		CLIENT SECTION
   ############################ */

func Client(port int) {
	sequence := rand.Uint32()
	var sequence_ptr *uint32 = &sequence
	// Establish a connection
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.Close()
	for {

		err := clientThreeWayHandshake(conn, sequence_ptr)
		if err != nil {
			fmt.Println("Error in handshake:", err)
		} else {
			break
		}
	}

}

func clientThreeWayHandshake(conn *net.UDPConn, sequence *uint32) error {

	// Send SYN packet
	err := sendSyn(conn, *sequence)
	if err != nil {
		return err
	}
	*sequence++

	recv_packet, err := receiveAndVerifySynAck(conn, *sequence)
	if err != nil {
		return err
	}

	err = clientSendAck(recv_packet, conn, *sequence)
	if err != nil {
		return err
	}
	*sequence++

	fmt.Println("Handshake completed with server.")
	return nil
}

func sendSyn(conn *net.UDPConn, sequence uint32) error {
	packet := &TCPPacket{
		flags:    SYN,
		sequence: sequence,
	}

	_, err := conn.Write(pack(packet))
	if err != nil {
		return err
	}
	fmt.Println("Sending SYN to: ", conn.LocalAddr().String())
	return nil
}

func receiveAndVerifySynAck(conn *net.UDPConn, sequence uint32) (*TCPPacket, error) {
	buf := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil || n == 0 {
		return nil, err
	}
	recv_packet := unpack(buf)
	if recv_packet.flags != SYN_ACK {
		return nil, fmt.Errorf("Wrong packet received")
	}

	fmt.Println("SYN_ACK received from: ", conn.LocalAddr().String())
	if recv_packet.acknowledgement != sequence {
		return nil, fmt.Errorf("SYN_ACK ack number did not match sequence")
	}
	return recv_packet, nil
}

func clientSendAck(recv_packet *TCPPacket, conn *net.UDPConn, sequence uint32) error {
	resp_packet := &TCPPacket{
		flags:           ACK,
		acknowledgement: recv_packet.sequence + 1,
		sequence:        sequence,
	}

	fmt.Println("sending ACK with acknowledgement")
	_, err := conn.Write(pack(resp_packet))
	if err != nil {
		return err
	}
	return nil
}
