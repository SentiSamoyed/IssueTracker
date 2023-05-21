package main

import (
	"log"
	"time"
)

type Answer int

const (
	UNDONE    Answer = 0
	DONE      Answer = 1
	EXISTED   Answer = 2
	NOT_EXIST Answer = 3
	FAILED    Answer = iota
)

type TrackerRequest struct {
	FullName string
	Chan     chan Answer
}

type TrackerResult struct {
	FullName string
	Answer   Answer
	Err      error
}

var reqChan chan TrackerRequest
var resChan chan TrackerResult

func InitTracker() {
	reqChan = make(chan TrackerRequest, 50)
	resChan = make(chan TrackerResult, 50)
	go mainLoop()
}

func TrackerSubmit(fullName string) Answer {
	ch := make(chan Answer, 1)
	reqChan <- TrackerRequest{
		FullName: fullName,
		Chan:     ch,
	}

	return <-ch
}

func mainLoop() {
	tasks := make(map[string]Answer, 0)

	for {
		select {
		case req := <-reqChan:
			fullName := req.FullName
			ch := req.Chan
			if ans, ex := tasks[fullName]; ex {
				// The task has been done before
				ch <- ans
			} else {
				tasks[fullName] = UNDONE
				go handleRequest(fullName)
				ch <- UNDONE
			}

			break
		case res := <-resChan:
			if res.Answer == DONE || res.Answer == EXISTED {
				tasks[res.FullName] = DONE
			} else {
				log.Printf("Error on %v: %v\n", res.FullName, res.Err.Error())
			}

			break
		default:
			break
		}
	}
}

func handleRequest(fullName string) {
	log.Printf("Received request on %v\n", fullName)
	time.Sleep(time.Second * 5)
	log.Printf("Request on %v is done\n", fullName)
	resChan <- TrackerResult{
		FullName: fullName,
		Answer:   DONE,
		Err:      nil,
	}
}
