package main

import "log"

func StartFollower(leader string)  {

	log.Printf("Follower started")
	log.Printf("Contacting leader at %s", leader)

}