package main

// Operation type mapping
// ADD - 0
// PENDING - 1
// DELETE - 2
// NOTHING - 3

/// Message type mapping
// REQ - 0
// OK - 1
// NEWVIEW - 2
// NEWLEADER - 3
type Message struct {
	Type int
	Data map[string]int
}

func ReqMessage(rid int, cid int, pid int, opType int) Message {

	m := make(map[string]int)
	m["reqId"] = rid
	m["curViewId"] = cid
	m["procId"] = pid
	m["opType"] = opType

	return Message{
		Type: 0, // REQ message
		Data: m,
	}
}

func IsReqMessage(message *Message) bool {
	return message.Type == 0
}

func IsOkMessage(message *Message) bool {
	return message.Type == 1
}

func IsNewViewMessage(message *Message) bool {
	return message.Type == 2
}

func IsNewLeaderMessage(message *Message) bool {
	return message.Type == 3
}

func AddReqMessage(rid int, cid int, pid int) Message {
	return ReqMessage(rid, cid, pid, 0)
}

func IsAddReqMessage(msg *Message) bool {
	return msg.Data["opType"] == 0
}

func DeleteReqMessage(rid int, cid int, pid int) Message {
	return ReqMessage(rid, cid, pid, 2)
}

func IsDeleteReqMessage(msg *Message) bool {
	return msg.Data["opType"] == 2
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

func NewViewMessage(cid int, mMap map[string]int) Message {

	mMap["curViewId"] = cid

	return Message{
		Type: 2,
		Data: mMap,
	}
}

func NewLeaderMessage(rid int, cid int) Message {
	m := make(map[string]int)
	m["reqId"] = rid
	m["curViewId"] = cid
	m["opType"] = 1

	return Message{
		Type: 3,
		Data: m,
	}
}

func IsPendingReqMessage(msg *Message) bool {
	return msg.Data["opType"] == 1
}

func IsNothingReqMessage(msg *Message) bool {
	return msg.Data["opType"] == 3
}
