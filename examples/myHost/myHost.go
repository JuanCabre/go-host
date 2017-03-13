package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	dbg "github.com/JuanCabre/go-debug"
	host "github.com/JuanCabre/go-host/src/host"
)

var debug = dbg.Debug("Main")

func main() {
	target := flag.String("target", "", "target")
	network := flag.String("network", "tcp", "udp or tcp")
	message := flag.String("message", "Hello World!", "message")
	flag.Parse()

	if *target == "" {
		h, err := host.NewHost("127.0.0.1")
		if err != nil {
			log.Fatal(err)
		}

		debug("Creating a udp service")
		err = h.NewService("udp", "echoUDP", "9100", doEchoGenPacket)
		if err != nil {
			log.Fatal(err)
		}
		debug("Creating a tcp service")
		err = h.NewService("tcp", "echoTCP", "9101", doEchoGenConn)
		if err != nil {
			log.Fatal(err)
		}

		debug("Creating a udp service")
		err = h.NewService("udp", "echoUDP", "9102", doEchoUDP)
		if err != nil {
			log.Fatal(err)
		}
		debug("Creating a tcp service")
		err = h.NewService("tcp", "echoTCP", "9103", doEchoTCP)
		if err != nil {
			log.Fatal(err)
		}

		for {
		}
	}
	// Here ends the listener

	conn, err := net.Dial(*network, *target)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write([]byte(*message))
	if err != nil {
		log.Fatal(err)
	}

	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response: ", string(response[:n]))
}

func doEchoTCP(conn *net.TCPConn) {
	debug("Calling doEchoTCP")
	payload := make([]byte, 512)
	n, err := conn.Read(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Received: ", string(payload[:n]), "Doing echo")

	conn.Write(payload[:n])
}

func doEchoUDP(conn *net.UDPConn) {
	debug("Calling doEchoUDP")
	payload := make([]byte, 512)
	n, addr, err := conn.ReadFromUDP(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Received: ", string(payload[:n]), "Doing echo Gen")

	_, err = conn.WriteToUDP(payload[:n], addr)
	if err != nil {
		log.Fatal(err)
	}

}

func doEchoGenPacket(conn net.PacketConn) {
	payload := make([]byte, 512)
	n, addr, err := conn.ReadFrom(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Received: ", string(payload[:n]), "Doing echo GenPacket ")

	_, err = conn.WriteTo(payload[:n], addr)
	if err != nil {
		log.Fatal(err)
	}

}

func doEchoGenConn(conn net.Conn) {
	payload := make([]byte, 512)
	n, err := conn.Read(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Received: ", string(payload[:n]), "Doing echo GenConn")

	_, err = conn.Write(payload[:n])
	if err != nil {
		log.Fatal(err)
	}

}
