package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

// Given a Listener and a channel, this function will receive data from the socket,
// decodes it to Message struct and send it through the channel. Used as a goroutine
// in Leader and follower event loops.
func acceptTCPMessages(tcp net.Listener, inMsgChan chan Message) {
	buf := make([]byte, 1024)

	for {
		conn, err := tcp.Accept()
		LogFatalCheck(err, "Cannot accept connection.")

		_, err = conn.Read(buf)
		LogFatalCheck(err, "Error reading from tcp socket")

		// Decode received Message
		msg := new(Message)
		gobj := gob.NewDecoder(bytes.NewBuffer(buf))
		gobj.Decode(msg)

		log.Printf("accpeted TCP message from %s", conn.RemoteAddr())

		inMsgChan <- *msg
	}

}

// Send Message to a process using TCP socket
func sendTCPMsg(m Message, toAddr string) {
	buf := new(bytes.Buffer)

	conn, err := net.Dial("tcp", toAddr)
	LogFatalCheck(err, fmt.Sprintf("Error establishing tcp connection to %s", toAddr))
	defer conn.Close()

	gobobj := gob.NewEncoder(buf)
	gobobj.Encode(m)

	conn.Write(buf.Bytes())
}

// Listen for heartbeats and send the hostname through channel
func acceptUDPHeartbeats(udp net.UDPConn, hbMsgChan chan string) {

	buf := make([]byte, 8)

	for {
		_, remoteAddr, _ := udp.ReadFromUDP(buf)
		sender, _ := net.LookupAddr(remoteAddr.IP.String())

		println("Got heartbeat from", sender[0])

	}
}

func sendHeartbeat(toAddr string) {
	conn, err := net.Dial("udp", toAddr)
	LogFatalCheck(err, fmt.Sprintf("Error establishing udp connection to %s", toAddr))
	defer conn.Close()

	conn.Write([]byte("ping"))
}
