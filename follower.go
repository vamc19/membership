package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"time"
)

func StartFollower() {

	inMsgChan := make(chan InMsgType)
	defer close(inMsgChan)

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	LogFatalCheck(err, "Failed to create TCP listener")
	defer tcp.Close()

	go acceptTCPMessages(tcp, inMsgChan)

	// UDP Messages for heartbeats. Starts on (TCP port + 1)
	hbMsgChan := make(chan string) // monitorUDPHeartbeats sends the remote ip address
	defer close(hbMsgChan)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: port + 1})
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
			followerProcessMessage(msg.Message, msg.From)
		case hb := <-hbMsgChan:
			processFollowerHeartbeat(ipHostMap[hb])
		case <-hbGCTimer:
			checkHeartbeatTimes()
		}
	}
}

func followerProcessMessage(message Message, remotehost string) {

	// Save request to reqList and send ok message
	if IsReqMessage(&message) {
		key := [2]int{message.Data["reqId"], message.Data["curViewId"]}
		reqList[key] = message

		msg := OkMessage(message.Data["reqId"], message.Data["curViewId"])
		go sendTCPMsg(msg, fmt.Sprintf("%s:%d", leader, port))
	}

	// got a new view message. Update the memebershipList and viewId
	if IsNewViewMessage(&message) {
		viewId = message.Data["curViewId"]

		// update membershiplist
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
		//println("Got new leader message from", remotehost)
	}
}

func processFollowerHeartbeat(remoteHost string) {
	//log.Printf("got heartbeat from %s", remoteHost)
	lastHeartBeat[remoteHost] = time.Now()
}

func printMembership() {
	var pids []int
	for h := range membershipList {
		pids = append(pids, hostPidMap[h])
	}

	sort.Ints(pids)
	str := fmt.Sprintf("Current View: %d. Members:", viewId)
	for i := range pids {
		str += " " + strconv.Itoa(i)
	}

	log.Printf(str)
}

func checkHeartbeatTimes() {
	n := time.Now()
	for h, t := range lastHeartBeat {
		if n.Sub(t) > (time.Duration(heartbeatFreq*2) * time.Second) {
			log.Printf("Lost host %s", h)
		}
	}
}
