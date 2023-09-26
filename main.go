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
	time.Sleep(1 * time.Second)
	go client(25554)
	go client(25555)
	time.Sleep(50 * time.Second)
}

/* ############################
		TCP SECTION
   ############################ */

type TCPPacket struct {
	flags           byte
	sequence        uint32
	acknowledgement uint32
	data            []byte
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

/* ############################
		CLIENT SECTION
   ############################ */

func client(port int) {
	sequence := rand.Uint32()
	var sequence_ptr *uint32 = &sequence

	// Establish a connection
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		fmt.Println("Client: Error dialing:", err)
		return
	}
	defer conn.Close()
	recv_sequence, err := clientThreeWayHandshake(conn, sequence_ptr)
	if err != nil {
		fmt.Println("Client: Error in handshake:", err)
		return
	}

	//clientSend(lastSequence, conn, sequence, FIN)
	//sequence++
	time.Sleep(5 * time.Second)
	clientCloseConnection(recv_sequence, conn, sequence_ptr)
}

func clientCloseConnection(recv_sequence uint32, conn *net.UDPConn, sequence_ptr *uint32) {
	println("Client: Sending FIN")
	clientSend(recv_sequence, conn, sequence_ptr, FIN)
	//recive ack
	recv_packet, err := receiveAndVerify(conn, *sequence_ptr, ACK)
	if err != nil {
		return
	}
	recv_sequence = recv_packet.sequence
	println("Client: Recived ACK from Server")
	//recive fin
	recv_packet, err = receiveAndVerify(conn, *sequence_ptr, FIN)
	if err != nil {
		return
	}
	recv_sequence = recv_packet.sequence
	println("Client: Recived FIN from Server")
	//send ack
	println("Client: Sending ACK")
	clientSend(recv_sequence, conn, sequence_ptr, ACK)
	//connection finshed
}

func clientThreeWayHandshake(conn *net.UDPConn, sequence *uint32) (uint32, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		// Send SYN packet
		fmt.Println("Client: Sending SYN to: ", conn.LocalAddr().String())
		err := sendSYN(conn, sequence)
		if err != nil {
			fmt.Println(err)
			continue
		}

		recv_packet, err := receiveAndVerify(conn, *sequence, SYN_ACK)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Client:  Send ACK To server.")
		err = clientSend(recv_packet.sequence, conn, sequence, ACK)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Client: Handshake completed with server.")
		return recv_packet.sequence, nil
	}
	return 0, fmt.Errorf("Client: Failed to handshake with server")
}

func sendSYN(conn *net.UDPConn, sequence *uint32) error {
	packet := &TCPPacket{
		flags:    SYN,
		sequence: *sequence,
	}
	*sequence++
	_, err := conn.Write(pack(packet))
	if err != nil {
		return err
	}
	return nil
}

func receiveAndVerify(conn *net.UDPConn, sequence uint32, flags byte) (*TCPPacket, error) {
	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil || n == 0 {
		return nil, err
	}
	recv_packet := unpack(buf)
	if recv_packet.flags != flags {
		return nil, fmt.Errorf("Client: Wrong packet received")
	}

	fmt.Println("Client: Correct flag(s) received from: ", conn.LocalAddr().String(), " : ", flags)
	if recv_packet.acknowledgement != sequence {
		return nil, fmt.Errorf("Client: Ack number did not match sequence")
	}
	return recv_packet, nil
}

// sends all
func clientSend(recv_sequence uint32, conn *net.UDPConn, sequence *uint32, flags byte) error {
	resp_packet := &TCPPacket{
		flags:           flags,
		acknowledgement: recv_sequence + 1,
		sequence:        *sequence,
	}
	*sequence++
	_, err := conn.Write(pack(resp_packet))
	if err != nil {
		return err
	}
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
		if recv_packet.flags == SYN {
			handleSyn(recv_packet, sequenceMap, addr, conn)
		} else if recv_packet.flags == ACK {
			handleAck(recv_packet, sequenceMap, addr, conn)
		} else if recv_packet.flags == FIN {
			handleFIN(recv_packet, sequenceMap, addr, conn)
		}
	}
}

func handleFIN(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {
	if recv_packet.acknowledgement != sequenceMap[addr.String()] {
		fmt.Println("Server: FIN does not match server sequence IP: " + addr.String())
	}
	println("Server: Recived FIN from client")
	//send ack
	println("Server: Sending ACK")
	serverSend(recv_packet, sequenceMap, addr, conn, ACK)
	//send own fin
	println("Server: Sending FIN")
	serverSend(recv_packet, sequenceMap, addr, conn, FIN)
}

func handleSyn(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {

	fmt.Println("Server: SYN received")
	sequenceMap[addr.String()] = rand.Uint32()
	serverSend(recv_packet, sequenceMap, addr, conn, SYN_ACK)
}

func serverSend(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn, flags byte) {
	resp_packet := &TCPPacket{
		flags:           flags,
		acknowledgement: recv_packet.sequence + 1,
		sequence:        sequenceMap[addr.String()],
	}
	_, err := conn.WriteToUDP(pack(resp_packet), addr)
	sequenceMap[addr.String()]++

	if err != nil {
		fmt.Println("Server: Error sending server packet:", err)
		return
	}
}

func handleAck(recv_packet *TCPPacket, sequenceMap map[string]uint32, addr *net.UDPAddr, conn *net.UDPConn) {

	if recv_packet.acknowledgement != sequenceMap[addr.String()] {
		fmt.Println("Server: Ack does not match server sequence IP: " + addr.String())
	}
	fmt.Println("Server: Received ACK from client: " + addr.String())
}
