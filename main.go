package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var hostPidMap = make(map[string]int) // maps hostname to pid
var pidHostMap = make(map[int]string) // maps pid to hostname

// Working with hostnames is surprisingly hard - socket connections only show IP
// Populate hostname to IP mappings for all hosts in the hostsfile
var ipHostMap = make(map[string]string) // maps remote ip addresses to hostname

var leader string
var port int
var heartbeatFreq int

var viewId int
var reqId int
var membershipList = make(map[string]bool)     // hostname : true
var reqList = make(map[[2]int]Message)         // {reqId, viewId} : Message
var lastHeartBeat = make(map[string]time.Time) // host : Time of last heartbeat

func main() {

	portPtr := flag.Int("p", 10000, "Port the process will be listening on for incoming messages")
	pausePtr := flag.Int("pause", 0, "Sleep for specified time after startup")
	hostfilePtr := flag.String("h", "hostfile", "Path to hostfile")

	flag.Parse()

	// Sleep before continuing. Use to delay start of follower
	time.Sleep(time.Duration(*pausePtr) * time.Second)

	// read host file
	hosts := getHostListFromFile(*hostfilePtr)
	leader = hosts[0]

	port = *portPtr // TCP port. UDP port is (port+1)
	heartbeatFreq = 5

	hostname, err := os.Hostname()
	LogFatalCheck(err, "Error retrieving hostname")

	for i, h := range hosts {
		hostPidMap[h] = i
		pidHostMap[i] = h
		ipAddrs, _ := net.LookupHost(h)
		ipHostMap[ipAddrs[0]] = h
	}

	membershipList[leader] = true // Add leader to membershipList

	if leader == hostname {
		StartLeader()
	} else {
		StartFollower()
	}
}

func getHostListFromFile(filePath string) []string {
	f, err := os.Open(filePath)
	LogFatalCheck(err, fmt.Sprintf("Cannot read hostfile at %s", filePath))
	defer f.Close()

	scanner := bufio.NewScanner(f)
	LogFatalCheck(scanner.Err(), "Error scanning file")

	var s []string

	for scanner.Scan() {
		s = append(s, scanner.Text())
	}

	return s
}
