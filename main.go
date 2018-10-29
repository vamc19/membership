package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"log"
)

func main() {

	portPtr := flag.Int("p", 10000, "Port the process will be listening on for incoming messages")
	hostfilePtr := flag.String("h", "hostfile", "Path to hostfile")

	flag.Parse()

	// read host file
	hosts := getHostListFromFile(*hostfilePtr)
	println(*portPtr)

	hostname, err := os.Hostname()
	LogFatalCheck(err, "Error retrieving hostname")

	if hosts[0] == hostname {
		log.Printf("Master: %s", hostname)

		pidMap := make(map[string]int)
		for i, h := range hosts {
			pidMap[h] = i
		}

		StartLeader(pidMap)
	} else {
		StartFollower(hosts[0])
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