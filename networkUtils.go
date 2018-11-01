package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Given a Listener and a channel, this function will receive data from the socket,
// decodes it to Message struct and send it through the channel. Used as a goroutine
// in Leader and follower event loops.
func acceptTCPMessages(tcp net.Listener, inMsgChan chan InMsgType) {
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

		remoteIP := strings.Split(conn.RemoteAddr().String(), ":")[0]

		inMsgChan <- InMsgType{From: remoteIP, Message: *msg}
	}

}

// Send Message to a process using TCP socket
func sendTCPMsg(m Message, toAddr string) {
	buf := new(bytes.Buffer)

	conn, err := net.Dial("tcp", toAddr)
	if err != nil {
		log.Printf("Error establishing tcp connection to %s: Reason: %v", toAddr, err)
		return
	}
	defer conn.Close()

	gobobj := gob.NewEncoder(buf)
	gobobj.Encode(m)

	conn.Write(buf.Bytes())
}

// Listen for heartbeats and send the hostname through channel
func monitorUDPHeartbeats(udp net.UDPConn, hbMsgChan chan string) {

	for {
		buf := make([]byte, 8)

		_, remoteAddr, _ := udp.ReadFromUDP(buf)
		hbMsgChan <- remoteAddr.IP.String()
	}
}

func sendHeartbeat(toAddr string) {
	conn, err := net.Dial("udp", toAddr)
	LogFatalCheck(err, fmt.Sprintf("Error establishing udp connection to %s", toAddr))
	defer conn.Close()

	conn.Write([]byte("ping"))
}

func multicastHeartbeats() {
	for h := range membershipList {
		addr := fmt.Sprintf("%s:%d", h, port+1)
		go sendHeartbeat(addr)
	}
}

func startHeartbeat(frequency int) {

	for {
		multicastHeartbeats()
		time.Sleep(time.Duration(frequency) * time.Second)
	}
}

func startHBGCTimer(timerMsgChan chan bool) {
	for {
		time.Sleep(time.Duration(3) * time.Second)
		timerMsgChan <- true
	}
}

// Send message to all hosts in membership list.
func multicastTCPMessage(msg Message, exceptHost string) {
	for h := range membershipList {
		if h == exceptHost { // ignore failed node
			continue
		}
		addr := fmt.Sprintf("%s:%d", h, port)
		go sendTCPMsg(msg, addr)
	}
}
