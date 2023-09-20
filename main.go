package main

import (
	"bufio"
	"fmt"
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

func listen(port int) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Print(err)
	}
	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Print(err)
	}
	for {
		var buf [512]byte
		_, addr, err := ln.ReadFromUDP(buf[0:])
		println(addr)
		if err != nil {
			// handle error
		}
		fmt.Println(string(buf[0:]))
	}

}
