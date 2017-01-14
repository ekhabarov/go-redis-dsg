package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
)

const MAX_RAND_PROCESS_TIME = 1000

type Message string

//Waits for new messages by BRPOP
func waiter(c redis.Conn, queue string, task chan<- Message) {
	fmt.Println("Waiting for messages.")
	for {
		msg, err := c.Do("BRPOP", queue, 0)
		if err != nil {
			log.Println("unable to get message from redis: ", err)
		}

		v, err := redis.Values(msg, err)
		PanicIf(err)

		s, err := redis.String(v[1], err)
		PanicIf(err)

		task <- Message(s)
	}
}

//Make primary work
func worker(wid int, tc <-chan Message, result chan<- finishedTask) {
	//fmt.Printf("Worker %d started.\n", wid)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	processTime := time.Duration(r.Intn(MAX_RAND_PROCESS_TIME))

	for t := range tc {
		time.Sleep(time.Millisecond * processTime)
		result <- finishedTask{processTime, t, wid}
	}
}
