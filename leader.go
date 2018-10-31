package main

import (
	"fmt"
	"net"
)

func StartLeader(pidMap map[string]int, port int) {

	println("Leader started")
	//hostname, err := os.Hostname()

	// TCP Message Channel
	inMsgChan := make(chan Message)
	defer close(inMsgChan)

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	LogFatalCheck(err, "Failed to create TCP listener")
	defer tcp.Close()

	go acceptTCPMessages(tcp, inMsgChan)

	// UDP Messages for heartbeats. Starts on (TCP port + 1)
	hbMsgChan := make(chan string) // acceptUDPHeartbeats sends the hostname back
	defer close(hbMsgChan)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: port + 1})
	LogFatalCheck(err, "Failed to create UDP listener")
	defer udp.Close()

	go acceptUDPHeartbeats(*udp, hbMsgChan)

	for {
		msg := <-inMsgChan
		println("got msg", msg.Type)
	}

}
