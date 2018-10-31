package main

import (
	"fmt"
	"net"
)

type InMsgType struct {
	From    string
	Message Message
}

var okList = make(map[[2]int]map[string]bool) // {reqId, viewId} : set of hosts which sent OK messages for the request

func StartLeader() {

	// TCP Message Channel
	inMsgChan := make(chan InMsgType)
	defer close(inMsgChan)

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	LogFatalCheck(err, "Failed to create TCP listener")
	defer tcp.Close()

	go acceptTCPMessages(tcp, inMsgChan)

	// UDP Messages for heartbeats. Starts on (TCP port + 1)
	hbMsgChan := make(chan string) // acceptUDPHeartbeats sends the remote ip address
	defer close(hbMsgChan)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: port + 1})
	LogFatalCheck(err, "Failed to create UDP listener")
	defer udp.Close()

	go acceptUDPHeartbeats(*udp, hbMsgChan)
	go startHeartbeat(5)

	for {

		select {
		case msg := <-inMsgChan:
			processMessage(msg.Message, ipHostMap[msg.From])
		case hb := <-hbMsgChan:
			processHeartbeat(ipHostMap[hb])
		}
	}
}

func processMessage(message Message, remotehost string) {
	// if message is OK
	if IsReqMessage(&message) {
		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leader, port))
	}

	if IsOkMessage(&message) {
		println("Got ok message from", remotehost)
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		if len(okList[key]) == 0 {
			okList[key] = make(map[string]bool)
		}

		okList[key][remotehost] = true // add sender to OkList

		if len(okList[key]) == len(membershipList) { // got confirmation from all peers. Finalize.
			finalizeRequest(key)
		}
	}

	if IsNewViewMessage(&message) {
		//println("Ignoring new view message from", remotehost)
	}

	if IsNewLeaderMessage(&message) {
		//println("Got new leader message from", remotehost)
	}
}

func processHeartbeat(remoteHost string) {
	//log.Printf("Processing heartbeat from %s", remoteHost)

	if !membershipList[remoteHost] { // New follower?

		// Send req message to membership list
		msg := AddReqMessage(reqId, viewId, hostPidMap[remoteHost])
		reqList[[2]int{reqId, viewId}] = msg

		for host := range membershipList {
			go sendTCPMsg(msg, fmt.Sprintf("%s:%d", host, port))
		}
	}
}

func finalizeRequest(key [2]int) {
	msg := reqList[key]
	if IsAddReqMessage(&msg) { // add to membershipList
		hostname := pidHostMap[msg.Data["procId"]]

		membershipList[hostname] = true
		viewId += 1

		membershipMap := make(map[string]int)
		for h := range membershipList {
			membershipMap[h] = hostPidMap[h]
		}
		nvMsg := NewViewMessage(viewId, membershipMap)
		multicastTCPMessage(nvMsg)
	}

}

func multicastTCPMessage(msg Message) {
	for h := range membershipList {
		addr := fmt.Sprintf("%s:%d", h, port)
		go sendTCPMsg(msg, addr)
	}
}
