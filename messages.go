package main

// Operation type mapping
// ADD - 0
// PENDING - 1
// DELETE - 2

/// Message type mapping
// REQ - 0
// OK - 1
// NEWVIEW - 2
// NEWLEADER - 3
type Message struct {
	Type int
	Data map[string]int
}

func ReqMessage(rid int, cid int, pid int) Message {

	m := make(map[string]int)
	m["reqId"] = rid
	m["curViewId"] = cid
	m["procId"] = pid

	return Message{
		Type: 0, // REQ message
		Data: m,
	}
}

func OkMessage(rid int, cid int) Message {
	m := make(map[string]int)
	m["reqId"] = rid
	m["curViewId"] = cid

	return Message{
		Type: 1,
		Data: m,
	}
}
