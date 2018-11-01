package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
)

// All the Pending requests to NewLeader go in here...
var newLeaderPending = make(map[[2]int]Message) // {reqId, viewId} : Message

func followerMessageProcessor(message Message, fromHost string) {

	// Save request to reqList and send ok message
	if IsReqMessage(&message) {
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}

		if key[1] == viewId { // Pending messages to new leaderHostname
			newLeaderPending[key] = message

			if len(newLeaderPending) == len(membershipList)-1 { // Got messages from remaining peers
				//restartPending()
			}
		}

		reqList[key] = message
		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leaderHostname, port))
	}

	// got a new view message. Update the memebershipList and viewId
	if IsNewViewMessage(&message) {
		viewId = message.Data["curViewId"]
		membershipList = make(map[string]bool)

		// update membershiplist and currentViewId
		for k, v := range message.Data {
			if k == "curViewId" {
				viewId = v
				continue
			}
			membershipList[k] = true
		}
		printMembership()
	}

	if IsNewLeaderMessage(&message) {
		delete(membershipList, leaderHostname)
		leaderHostname = fromHost // update leaderHostname hostname
		sendPendingMessages()     // New leaderHostname will not have any pending messages anyway.
	}
}

// Sort and print membership info.
func printMembership() {
	var pids []int
	for h := range membershipList {
		pids = append(pids, hostPidMap[h])
	}
	sort.Ints(pids)
	str := fmt.Sprintf("Current View: %d. Members:", viewId)
	for _, p := range pids {
		str += " " + strconv.Itoa(p)
	}
	log.Printf(str)
}

func iAmNewLeader() bool {
	var pids []int
	hostname, _ := os.Hostname()

	for h := range membershipList {
		if h != leaderHostname {
			pids = append(pids, hostPidMap[h])
		}
	}

	sort.Ints(pids)
	return pidHostMap[pids[0]] == hostname
}

func startNewLeaderProtocol() {
	msg := NewLeaderMessage(reqId, viewId)

	for h := range membershipList {
		if !lostHosts[h] {
			addr := fmt.Sprintf("%s:%d", h, port)
			go sendTCPMsg(msg, addr)
		}
	}

	reqId += 1
}

func sendPendingMessages() {
	pending := false
	var msg Message

	for k, v := range reqList {
		if k[1] > viewId { // ViewId in future. Pending operation
			pending = true

			if IsAddReqMessage(&v) {
				msg = AddReqMessage(reqId, viewId, v.Data["procId"])
			}

			if IsDeleteReqMessage(&v) {
				msg = DeleteReqMessage(reqId, viewId, v.Data["procId"])
			}
		}
	}

	if !pending {
		msg = ReqMessage(reqId, viewId, 0, 3) // Nothing message
	}

	for h := range membershipList {
		addr := fmt.Sprintf("%s:%d", h, port)
		go sendTCPMsg(msg, addr)
	}
}
