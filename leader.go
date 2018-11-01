package main

import (
	"fmt"
	"log"
	"sort"
	"time"
)

type InMsgType struct {
	From    string
	Message Message
}

var okList = make(map[[2]int]map[string]bool) // {reqId, viewId} : set of hosts which sent OK messages for the request

// process message based on message type
func leaderMessageProcessor(message Message, fromHost string) {

	if IsReqMessage(&message) {
		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leaderHostname, port))
	}

	if IsOkMessage(&message) {
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		if len(okList[key]) == 0 {
			okList[key] = make(map[string]bool)
		}

		okList[key][fromHost] = true // add sender to OkList

		msg := reqList[key]
		// If ADD message, make sure all the hosts sent OK responses
		// If DELETE message, make sure all but the lost host sent OK responses
		if (IsAddReqMessage(&msg) && len(okList[key]) == len(membershipList)) || (IsDeleteReqMessage(&msg) && len(okList[key]) == len(membershipList)-1) {
			multicastNewViewMessage(key)
		}
	}

	if IsNewViewMessage(&message) {
		printMembership()
	}

}

// Got all OK messages. Build and send updated membership
func multicastNewViewMessage(key [2]int) {
	msg := reqList[key] // request in the temp area
	remotehost := pidHostMap[msg.Data["procId"]]

	if IsAddReqMessage(&msg) { // add to membershipList
		membershipList[remotehost] = true
		viewId += 1
	}

	if IsDeleteReqMessage(&msg) { // delete from membership
		viewId += 1
		delete(membershipList, remotehost)
		delete(lastHeartbeat, remotehost)
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

// Check if the heartbeat is from a new host.
// If yes, start adding host to membership. Finally update timestamp.
func leaderHeartbeatProcessor(remoteHost string) {

	if !membershipList[remoteHost] { // New follower?
		msg := AddReqMessage(reqId, viewId, hostPidMap[remoteHost]) // ADD REQ message to members
		reqList[[2]int{reqId, viewId}] = msg
		multicastTCPMessage(msg, "")
		reqId += 1
	}

	lastHeartbeat[remoteHost] = time.Now()
}

// Remove unreachable hosts
func removeFailedHost(failedHost string) {
	if lostHosts[failedHost] { // Already processing
		return
	}

	// Delete host from membership
	msg := DeleteReqMessage(reqId, viewId, hostPidMap[failedHost])
	reqList[[2]int{reqId, viewId}] = msg

	if failDuringRemove {
		leaderFailureScenario(failedHost, msg)
	}

	multicastTCPMessage(msg, failedHost)
	reqId += 1
	lostHosts[failedHost] = true
}

func leaderFailureScenario(failedHost string, message Message) {
	var pids []int

	for h := range membershipList {
		if (h != leaderHostname) && (h != failedHost) {
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
