package main

import (
	"fmt"
	"math/rand"
	"net"
)

/* ############################
		SERVER SECTION
   ############################ */

func Server(port int) {
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

		if !verifySequence(recv_packet, sequenceMap[addr.String()]) {
			continue // drop packet
		}
		if recv_packet.flags == SYN {
			handleSyn(recv_packet, sequenceMap, addr, conn)
		} else if recv_packet.flags == ACK {
			handleAck(addr)
		} else {
			handleDataPacket(recv_packet)
		}
		sequenceMap[addr.String()]++
	}
}

func verifySequence(recv_packet *TCPPacket, sequenceNr uint32) bool {
	if recv_packet.acknowledgement != sequenceNr {
		fmt.Println("Sequence did not match")
		return false
	}
	return true
}

func handleDataPacket(recv_packet *TCPPacket) {

}

func handleSyn(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {
	fmt.Println("SYN received")
	sequenceMap[addr.String()] = rand.Uint32()
	resp_packet := &TCPPacket{
		flags:           SYN_ACK,
		acknowledgement: recv_packet.sequence + 1,
		sequence:        sequenceMap[addr.String()],
	}
	_, err := conn.WriteToUDP(pack(resp_packet), addr)
	if err != nil {
		fmt.Println("Error sending SYN_ACK:", err)
		return
	}
}

func handleAck(addr *net.UDPAddr) {
	fmt.Println("Connection established with IP: " + addr.String())
}
