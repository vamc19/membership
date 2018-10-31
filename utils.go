package main

import (
	"errors"
	"fmt"
	"log"
)

func LogFatalCheck(e error, msg string) {
	if e != nil {
		log.Fatal(errors.New(fmt.Sprintf("%s: %v", msg, e)))
	}
}

func LogCheck(e error, msg string) {
	if e != nil {
		log.Printf(msg)
	}
}
