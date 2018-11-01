package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func StartEventLoop() {

	if failDuringRemove {
		log.Printf("Will execute testcase 4")
	}

	// TCP Message Channel
	inMsgChan := make(chan InMsgType)
	defer close(inMsgChan)

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	LogFatalCheck(err, "Failed to create TCP listener")
	defer tcp.Close()

	go acceptTCPMessages(tcp, inMsgChan)

	// UDP heartbeats message channel
	hbMsgChan := make(chan string) // monitorUDPHeartbeats sends the remote ip address
	defer close(hbMsgChan)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: port + 1}) //on (TCP port + 1)
	LogFatalCheck(err, "Failed to create UDP listener")
	defer udp.Close()

	hbGCTimer := make(chan bool)
	defer close(hbGCTimer)

	go monitorUDPHeartbeats(*udp, hbMsgChan) // Heartbeat monitor
	go startHeartbeat(heartbeatFreq)         // Heartbeat thread
	go startHBGCTimer(hbGCTimer)             // Heartbeat Garbage Collection timer

	for {
		select {
		case msg := <-inMsgChan:
			processMessage(msg.Message, ipHostMap[msg.From])
		case hb := <-hbMsgChan:
			processHeartbeat(ipHostMap[hb])
		case <-hbGCTimer:
			garbageCollectHeartbeats()
		}
	}
}

// For every TCP message received
func processMessage(msg Message, fromHost string) {
	if isLeader {
		leaderMessageProcessor(msg, fromHost)
	} else {
		followerMessageProcessor(msg, fromHost)
	}
}

// For every heartbeat message received
func processHeartbeat(fromHost string) {
	if isLeader {
		leaderHeartbeatProcessor(fromHost) // Have to add host if not already in membershipList
	} else {
		lastHeartbeat[fromHost] = time.Now() // Just update the timestamp
	}
}

// Garbage collect heartbeats periodically and detect failed processes.
func garbageCollectHeartbeats() {
	n := time.Now()
	for h, t := range lastHeartbeat {
		if n.Sub(t) > (time.Duration(heartbeatFreq*2) * time.Second) { // Message delay detected

			// If leader deleted host from membership, remove it from history
			if !membershipList[h] {
				delete(lastHeartbeat, h)
				delete(lostHosts, h)
				continue
			}

			log.Printf("Peer %d not reachable", hostPidMap[h])

			if isLeader && removeFailed {
				removeFailedHost(h)
			}
		}
	}
}
