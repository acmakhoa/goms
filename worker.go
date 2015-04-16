package main

import (
	"log"
	"time"
)

const (
	SMSPending   = iota // 0
	SMSProcessed        // 1
	SMSError            // 2
)

type SMS struct {
	UUID    string `json:"uuid"`
	Mobile  string `json:"mobile"`
	Body    string `json:"body"`
	Status  int    `json:"status"`
	Retries int    `json:"retries"`
	Device  string `json:"device"`
}

//TODO: should be configurable
const SMSRetryLimit = 3

var messages chan SMS
var wakeupMessageLoader chan bool

var bufferMaxSize int
var bufferLowCount int
var messageCountSinceLastWakeup int
var timeOfLastWakeup time.Time
var messageLoaderTimeout time.Duration
var messageLoaderCountout int
var messageLoaderLongTimeout time.Duration

func InitWorker(bufferSize, bufferLow, loaderTimeout, countOut, loaderLongTimeout int) {
	log.Println("--- InitWorker")

	bufferMaxSize = bufferSize
	bufferLowCount = bufferLow
	messageLoaderTimeout = time.Duration(loaderTimeout) * time.Minute
	messageLoaderCountout = countOut
	messageLoaderLongTimeout = time.Duration(loaderLongTimeout) * time.Minute

	messages = make(chan SMS, bufferMaxSize)
	log.Println("--- make new chanel message")
	go messageLoader(bufferMaxSize, bufferLowCount)
}


func EnqueueMessage(message *SMS,insertToDB bool) {
	log.Println("--- EnqueueMessage: ", message)
	if insertToDB {
		insertMessage(message)
	}
	//wakeup message loader and exit
	go func() {
		//notify the message loader only if its been to too long
		//or too many messages since last notification
		messageCountSinceLastWakeup++
		if messageCountSinceLastWakeup > messageLoaderCountout || time.Now().Sub(timeOfLastWakeup) > messageLoaderTimeout {
			log.Println("EnqueueMessage: ", "waking up message loader")
			wakeupMessageLoader <- true
			messageCountSinceLastWakeup = 0
			timeOfLastWakeup = time.Now()
		}
		log.Println("EnqueueMessage - anon: count since last wakeup: ", messageCountSinceLastWakeup)
	}()
}


func messageLoader(bufferSize, minFill int) {
	// Load pending messages from database as needed
	for {

		/*
		   - set a fairly long timeout for wakeup
		   - if there are very few number of messages in the system and they failed at first go,
		   and there are no events happening to call EnqueueMessage, those messages might get
		   stalled in the system until someone knocks on the API door
		   - we can afford a really long polling in this case
		*/
	

		log.Println("messageLoader: start receive message")
		message := <-messages
		log.Println("messageLoader: receive message :",message)

		
	}
}

func AddMessage(message SMS) {
	log.Println("--- AddMessage: ", message)
	
	//wakeup message loader and exit
	go func() {
		//notify the message loader only if its been to too long
		//or too many messages since last notification
		messages <- message
		log.Println("AddMessage :Add message to channel")
	}()
}
