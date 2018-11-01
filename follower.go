package main

import (
	"fmt"
)

func followerMessageProcessor(message Message, fromHost string) {

	// Save request to reqList and send ok message
	if IsReqMessage(&message) {
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		reqList[key] = message

		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leaderHostname, port))
	}

	// got a new view message. Update the memebershipList and viewId
	if IsNewViewMessage(&message) {
		viewId = message.Data["curViewId"]
		membershipList = make(map[string]bool) // New membership list

		// update membershiplist and currentViewId
		for k, v := range message.Data {
			if k == "curViewId" {
				viewId = v
				continue
			}
			membershipList[k] = true
		}

		if justJoined {
			go multicastHeartbeats() // If just joined, let everyone know you are alive.
			justJoined = false
		}
		printMembership()
	}

	if IsNewLeaderMessage(&message) {
		// It is possible that this host has not yet detected that the leader is down yet, but new leader has.
		// If this message is received, just replace the old leader with new.
		if leaderHostname != fromHost {
			deleteMember(leaderHostname)
			leaderHostname = findNewLeader() // Should be the same as fromHost anyway.
		}

		sendPendingMessages() // New leaderHostname will not have any pending messages anyway.
	}
}

func sendPendingMessages() {
	pending := false
	var msg Message

	for k, v := range reqList {
		if k[1] == viewId { // View Id is not yet updated. Probably pending?
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

	// Send it to only leader
	go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leaderHostname, port))
}
