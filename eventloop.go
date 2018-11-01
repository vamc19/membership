package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
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

	for h := range membershipList {
		t, ok := lastHeartbeat[h]

		if !ok { // Maybe a new member, wait for it to ping
			continue
		}

		if n.Sub(t) > (time.Duration(heartbeatFreq*2) * time.Second) {

			log.Printf("Peer %d not reachable", hostPidMap[h])

			// Is leader, not executing test case 2 and is not a recovering leader
			if isLeader && removeFailedFlag && !recoveringLeader {
				removeFailedHost(h)
			}

			// if leader is down...
			if h == leaderHostname {
				// Delete old leader from membership list
				deleteMember(leaderHostname)

				// Find new leader and assign it
				leaderHostname = findNewLeader()
				log.Printf("Leader unavailable. New leader is %s", leaderHostname)

				// Am I the new leader?
				if leaderHostname == hostname {
					isLeader = true
					failDuringRemove = false // Flag did its part, relieve it of its duty
					recoveringLeader = true
					startNewLeaderProtocol()
				}
			}
		}
	}
}

// Remove all the traces of a lost host
func deleteMember(host string) {
	delete(membershipList, host)
	delete(lastHeartbeat, host)
	delete(tempLostList, host)
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

// Find the hostname of next leader
func findNewLeader() string {
	var pids []int

	for h := range membershipList {
		if h != leaderHostname {
			pids = append(pids, hostPidMap[h])
		}
	}

	sort.Ints(pids)
	return pidHostMap[pids[0]]
}

// Send NEWLEADER message
func startNewLeaderProtocol() {
	msg := NewLeaderMessage(reqId, viewId)

	for h := range membershipList {
		if !tempLostList[h] { // There may be hosts that are already lost. Ignore them.
			addr := fmt.Sprintf("%s:%d", h, port)
			go sendTCPMsg(msg, addr)
		}
	}

	reqId += 1
}
