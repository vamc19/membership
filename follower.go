package main

import (
	"fmt"
	"log"
	"time"
)

func StartFollower(leader string, port int) {

	log.Printf("Follower started")
	log.Printf("Contacting leader at %s", leader)

	leaderTCPAddr := fmt.Sprintf("%s:%d", leader, port)
	leaderUDPAddr := fmt.Sprintf("%s:%d", leader, port+1)

	inMsgChan := make(chan Message)
	defer close(inMsgChan)

	sendHeartbeat(leaderUDPAddr)

	// lets create the message we want to send accross
	msg := OkMessage(1, 1)
	msg2 := OkMessage(2, 3)

	sendTCPMsg(msg, leaderTCPAddr)
	sendTCPMsg(msg2, leaderTCPAddr)

	time.Sleep(2 * time.Second)
}
