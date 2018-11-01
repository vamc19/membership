package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"time"
)

type InMsgType struct {
	From    string
	Message Message
}

var okList = make(map[[2]int]map[string]bool) // {reqId, viewId} : set of hosts which sent OK messages for the request

func StartLeader(failDuringRemove bool) {

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

	go monitorUDPHeartbeats(*udp, hbMsgChan)
	go startHeartbeat(heartbeatFreq)
	go startHBGCTimer(hbGCTimer)

	for {

		select {
		case msg := <-inMsgChan:
			processMessage(msg.Message, ipHostMap[msg.From])
		case hb := <-hbMsgChan:
			processHeartbeat(ipHostMap[hb])
		case <-hbGCTimer:
			checkHeartbeatAndRemove(failDuringRemove)
		}
	}
}

// process message based on type
func processMessage(message Message, remotehost string) {

	if IsReqMessage(&message) {
		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leader, port))
	}

	if IsOkMessage(&message) {
		//println("Got ok message from", remotehost)
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		if len(okList[key]) == 0 {
			okList[key] = make(map[string]bool)
		}

		okList[key][remotehost] = true // add sender to OkList

		msg := reqList[key]
		if (IsAddReqMessage(&msg) && len(okList[key]) == len(membershipList)) || (IsDeleteReqMessage(&msg) && len(okList[key]) == len(membershipList)-1) {
			finalizeRequest(key)
		}
	}

	if IsNewViewMessage(&message) {
		printMembership()
	}

}

// Check if the heartbeat is from a new host.
// If yes, start adding host to membership. Finally update timestamp.
func processHeartbeat(remoteHost string) {

	if !membershipList[remoteHost] { // New follower?
		msg := AddReqMessage(reqId, viewId, hostPidMap[remoteHost]) // ADD REQ message to members
		reqList[[2]int{reqId, viewId}] = msg
		multicastTCPMessage(msg, "")
		reqId += 1
	}

	lastHeartBeat[remoteHost] = time.Now()
}

// Got all OK messages. Finalize pending request
func finalizeRequest(key [2]int) {
	msg := reqList[key] // request in the temp area
	remotehost := pidHostMap[msg.Data["procId"]]

	if IsAddReqMessage(&msg) { // add to membershipList
		membershipList[remotehost] = true
		viewId += 1
	}

	if IsDeleteReqMessage(&msg) { // delete from memebership
		viewId += 1
		delete(membershipList, remotehost)
		delete(lastHeartBeat, remotehost)
		delete(lostHosts, remotehost)
	}

	// Broadcast new membership list
	currentMembers := make(map[string]int)
	for h := range membershipList {
		currentMembers[h] = hostPidMap[h]
	}
	nvMsg := NewViewMessage(viewId, currentMembers)
	multicastTCPMessage(nvMsg, "")
}

// send message to all hosts in membership list.
func multicastTCPMessage(msg Message, exceptHost string) {
	for h := range membershipList {
		if h == exceptHost { // ignore failed node
			continue
		}
		addr := fmt.Sprintf("%s:%d", h, port)
		go sendTCPMsg(msg, addr)
	}
}

func checkHeartbeatAndRemove(failDuringRemove bool) {
	n := time.Now()
	for h, t := range lastHeartBeat {
		if lostHosts[h] { // Already processing
			continue
		}

		if n.Sub(t) > (time.Duration(heartbeatFreq*2) * time.Second) {
			log.Printf("Peer %d not reachable", hostPidMap[h])

			// Delete host from membership
			msg := DeleteReqMessage(reqId, viewId, hostPidMap[h])
			reqList[[2]int{reqId, viewId}] = msg

			if failDuringRemove {
				leaderFailure(h, msg)
			}

			multicastTCPMessage(msg, h)
			reqId += 1
			lostHosts[h] = true
		}
	}
}

func leaderFailure(failedHost string, message Message) {
	var pids []int

	for h := range membershipList {
		if (h != leader) && (h != failedHost) {
			pids = append(pids, hostPidMap[h])
		}
	}

	sort.Ints(pids)
	nextLeader := pidHostMap[pids[0]]

	for h := range membershipList {
		if h == nextLeader || h == failedHost { // ignore failed node
			continue
		}
		addr := fmt.Sprintf("%s:%d", h, port)
		go sendTCPMsg(message, addr)
	}

	log.Fatal("So long and thanks for all the fish!")
}
