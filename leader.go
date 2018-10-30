package main

import (
	"bytes"
	"fmt"
	"net"
)

func StartLeader(pidMap map[string]int, port int) {

	println("Leader started")

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	LogFatalCheck(err, "Failed to create server")
	defer tcp.Close()

	conn, err := tcp.Accept()
	LogFatalCheck(err, "Failed to accept connection")
	tmp := make([]byte, 500)

	for {
		_, err = conn.Read(tmp)
		tmpbuff := bytes.NewBuffer(tmp)

	}

}
