package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
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
	conn, err := net.Dial("udp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		// handle error
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Print(status)
}


const (
	SYN  byte = 0b00000001
	ACK  byte = 0b00000010
	FIN  byte = 0b00000100
	DATA byte = 0b00001000
	SYN_ACK byte = 0b00000011
)


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
		n, addr, err := ln.ReadFromUDP(buf[0:])
		println(addr)
		println(n)
		if err != nil {
			// handle error
		}
		flags := buf[0] // assuming the first byte contains the flags
		state := ConnectionsState.GetState(addr)
		//first check the flag bits
		switch{
		//if first syn is recevied return syn + ack
		case flags&SYN == SYN:
			if state == ""{
				ConnectionsState.SetState(addr, StateAwatingACK)
				acknowledgementINT := rand.Uint32()
				acknowledgementMap[addr.String()] = acknowledgementINT

				//convert into 4 bytes
				acknowledgementBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(acknowledgementBytes, uint32(acknowledgementINT))
				header := make([]byte, 5)
				//add the byte for syn and ack, then add the integer for acknowlegement into a header
				header[0] = SYN_ACK
				for i:=1; i<5; i++{
					header[i] = acknowledgementBytes[i-1]
				}
				ln.WriteToUDP(header, addr)
			}
		//wait next falg should allways be ACK for the same incoming addr
		case flags&ACK == ACK:
			if state == StateAwatingACK{
				acknowledgementBytes := make([]byte, 4)
				for i := 1; i < 5 ; i++{
					acknowledgementBytes[i-1] = buf[i]
				}					
				acknowledgementUINT := binary.BigEndian.Uint32(acknowledgementBytes)
				if acknowledgementMap[addr.String()] == acknowledgementUINT{
					ConnectionsState.SetState(addr, StateConnectionEstablished)
				} 
				//TODO: if the acknowledgement number is not the same handle
			}
			if state == StateAwatingFinACK{

			}
		//now date is ready to be resived after each recived data block return empty responce with flag bit ACK activated
		

		

		//if fin is recived unless data is missing return ACK

		//then return FIN and wait for ACK

		//now the connection is closed
		}
	}
}


//we want to ensure that multible messages can be sent a once 
type ConnectionState string

const (
	StateAwatingACK  ConnectionState = "AwaitACK"
	StateConnectionEstablished ConnectionState = "ConnectionEstablished"
	StateAwatingFinACK ConnectionState = "AwatingACKForFIN"
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
