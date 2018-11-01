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

// All the Pending requests to NewLeader go in here...
var pendingMessages []Message

// process message based on message type
func leaderMessageProcessor(message Message, fromHost string) {

	if IsReqMessage(&message) {
		if recoveringLeader {
			pendingMessages = append(pendingMessages, message)

			// Check if received messages from all alive hosts
			// 2 because not considering itself and lost leader.
			// tempLost list may contain a lost host which is not removed before leader failed
			if len(pendingMessages) == (len(membershipList) - 2 - len(tempLostList)) {
				restartPendingProtocol()
			}

			return
		}

		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leaderHostname, port))
	}

	if IsOkMessage(&message) {
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		if len(okList[key]) == 0 {
			okList[key] = make(map[string]bool)
		}

		okList[key][fromHost] = true // add sender to OkList

		sentMsg := reqList[key]
		// If ADD message, make sure all the hosts sent OK responses
		// If DELETE message, make sure all but the lost host sent OK responses
		if (IsAddReqMessage(&sentMsg) && len(okList[key]) == len(membershipList)) || (IsDeleteReqMessage(&sentMsg) && len(okList[key]) == len(membershipList)-1) {
			multicastNewViewMessage(key)
		}
	}

	if IsNewViewMessage(&message) {
		printMembership()
	}
}

// Start protocol to add a host to membership list
func startAddOperation(host string) {
	msg := AddReqMessage(reqId, viewId, hostPidMap[host]) // ADD REQ message to members
	reqList[[2]int{reqId, viewId}] = msg
	multicastTCPMessage(msg, "")
	reqId += 1
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
		deleteMember(remotehost)
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

	if !membershipList[remoteHost] && !removedList[remoteHost] { // New follower?
		startAddOperation(remoteHost)
	}

	lastHeartbeat[remoteHost] = time.Now()
}

// Remove unreachable hosts
func removeFailedHost(failedHost string) {
	if tempLostList[failedHost] { // Already processing
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
	tempLostList[failedHost] = true
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

	time.Sleep(1 * time.Second) // Stay alive until the messages are sent
	log.Fatal("So long and thanks for all the fish!")
}

func restartPendingProtocol() {
	var pendingMessage Message

	for _, m := range pendingMessages {
		if IsNothingReqMessage(&m) {
			continue
		}

		pendingMessage = m
		break
	}

	recoveringLeader = false
	if IsAddReqMessage(&pendingMessage) {
		log.Printf("Does not support ADD operation while leader restart")
		return
	}

	removeFailedHost(pidHostMap[pendingMessage.Data["procId"]])
}
