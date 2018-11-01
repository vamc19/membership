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

var leaderHostname string
var hostname string
var port int
var heartbeatFreq int
var isLeader bool
var failDuringRemove bool // flag for test case 4
var removeFailedFlag bool // flag for test case 2
var recoveringLeader bool
var justJoined bool

var viewId int
var reqId int
var membershipList = make(map[string]bool)     // hostname : true
var reqList = make(map[[2]int]Message)         // {reqId, viewId} : Message
var lastHeartbeat = make(map[string]time.Time) // host : Time of last heartbeat
var tempLostList = make(map[string]bool)       // all the hosts that are no longer alive
var removedList = make(map[string]bool)        // all the hosts that are removed from membership

func main() {

	portFlag := flag.Int("p", 10000, "Port the process will be listening on for incoming messages")
	pauseFlag := flag.Int("pause", 0, "Sleep for specified time after startup")
	leaderFailFlag := flag.Bool("t4", false, "Simulate test case 4")
	persistFailNodeFlag := flag.Bool("t2", false, "Simulate test case 2. Do not remove failed nodes")
	hostfileFlag := flag.String("h", "hostfile", "Path to hostfile")

	flag.Parse()

	// Sleep before continuing. Use to delay start of follower
	time.Sleep(time.Duration(*pauseFlag) * time.Second)

	// read host file
	hosts := getHostListFromFile(*hostfileFlag)
	leaderHostname = hosts[0]

	port = *portFlag // TCP port. UDP port is (port+1)
	heartbeatFreq = 5

	// Set test case flags
	failDuringRemove = *leaderFailFlag
	removeFailedFlag = !*persistFailNodeFlag
	recoveringLeader = false
	justJoined = true

	hostname, _ = os.Hostname()

	for i, h := range hosts {
		pid := i + 1
		hostPidMap[h] = pid
		pidHostMap[pid] = h
		ipAddrs, _ := net.LookupHost(h)
		ipHostMap[ipAddrs[0]] = h
	}

	membershipList[leaderHostname] = true // Add leaderHostname to membershipList
	isLeader = leaderHostname == hostname

	StartEventLoop()
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
