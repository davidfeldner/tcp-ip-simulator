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
	//if any bytes are needed it is added here
)

func main() {
	go listen(25555)
	go listen(25554)
	time.Sleep(4 * time.Second)
	go send(25555)
	go send(25554)
	time.Sleep(50 * time.Second)

}

func send(port int) {
	// Establish a connection
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}

	defer conn.Close()
	// Send SYN packet
	_, err = conn.Write([]byte{SYN})
	if err != nil {
		fmt.Println("Error sending SYN:", err)
		return
	}
	fmt.Println("Sending SYN")
	// Await SYN-ACK response
	buf := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error reading SYN-ACK:", err)
		return
	}
	if n == 0 {
		return
	}
	flags := buf[0]
	if flags&SYN_ACK == SYN_ACK {
		fmt.Println("SYN_ACK received")
		header := make([]byte, 5)
		header[0] = ACK
		//acknowledgement shoud be in 1-4 in the buffer
		for i := 1; i < 5; i++ {
			header[i] = buf[i]
		}
		fmt.Println("sending ACK with acknowledgement")
		_, err = conn.Write(header)
		if err != nil {
			fmt.Println("Error sending ACK:", err)
			return
		}
		fmt.Println("Handshake completed with server.")
	}
}

func listen(port int) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Print(err)
	}
	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Print(err)
	}
	ConnectionsState := NewStateMachine()
	acknowledgementMap := make(map[string]uint32)

	for {
		var buf [512]byte
		_, addr, err := ln.ReadFromUDP(buf[0:])
		if err != nil {
			// handle error
		}
		flags := buf[0] // assuming the first byte contains the flags
		state := ConnectionsState.GetState(addr)
		//first check the flag bits
		switch {
		//if first syn is recevied return syn + ack
		case flags&SYN == SYN:
			if state == "" {
				fmt.Println("SYN received")
				ConnectionsState.SetState(addr, StateAwatingACK)
				acknowledgementINT := rand.Uint32()
				acknowledgementMap[addr.String()] = acknowledgementINT

				//convert into 4 bytes
				acknowledgementBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(acknowledgementBytes, uint32(acknowledgementINT))
				header := make([]byte, 5)
				//add the byte for syn and ack, then add the integer for acknowlegement into a header
				header[0] = SYN_ACK
				for i := 1; i < 5; i++ {
					header[i] = acknowledgementBytes[i-1]
				}
				fmt.Println("sending SYN_ACK with acknowledgement")
				ln.WriteToUDP(header, addr)
			}
		//wait next falg should allways be ACK for the same incoming addr
		case flags&ACK == ACK:
			if state == StateAwatingACK {
				acknowledgementBytes := make([]byte, 4)
				for i := 1; i < 5; i++ {
					acknowledgementBytes[i-1] = buf[i]
				}
				acknowledgementUINT := binary.BigEndian.Uint32(acknowledgementBytes)
				if acknowledgementMap[addr.String()] == acknowledgementUINT {
					fmt.Println("ACK with correct acknowledgement received now ready for data transfer")
					ConnectionsState.SetState(addr, StateConnectionEstablished)
				}
				//TODO: if the acknowledgement number is not the same handle
			}
			if state == StateAwatingFinACK {

			}
			//now date is ready to be resived after each recived data block return empty responce with flag bit ACK activated

			//if fin is recived unless data is missing return ACK

			//then return FIN and wait for ACK

			//now the connection is closed
		}
	}
}

// we want to ensure that multible messages can be sent a once
// only to be used for the server
type ConnectionState string

const (
	StateAwatingACK            ConnectionState = "AwaitACK"
	StateConnectionEstablished ConnectionState = "ConnectionEstablished"
	StateAwatingFinACK         ConnectionState = "AwatingACKForFIN"
)

type StateMachine struct {
	ConnectionStates map[string]ConnectionState
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		ConnectionStates: make(map[string]ConnectionState),
	}
}

func (sm *StateMachine) SetState(addr *net.UDPAddr, state ConnectionState) {
	sm.ConnectionStates[addr.String()] = state
}

func (sm *StateMachine) GetState(addr *net.UDPAddr) ConnectionState {
	return sm.ConnectionStates[addr.String()]
}
