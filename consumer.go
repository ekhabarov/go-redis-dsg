package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	//Max number or ms for processing message
	MAX_RAND_PROCESS_TIME = 1000

	//Command for stoping workers
	CTRL_STOP = iota
)

type (
	Message string

	BadMessage struct {
		msg Message
		err string
	}

	ProcessedMessage struct {
		duration time.Duration
		msg      Message
		worker   int
	}

	Consumer struct {
		pool          *redis.Pool
		queue         string
		errQueue      string
		in            chan Message
		out           chan ProcessedMessage
		bad           chan BadMessage
		maxGoroutines int
		control       chan byte
	}
)

func (b *BadMessage) String() string {
	return fmt.Sprintf("m:%q e:%q", b.msg, b.err)
}

//Create new Consumer struct
func NewConsumer(p *redis.Pool, q string, eq string, mg int) *Consumer {
	return &Consumer{
		pool:          p,
		queue:         q,
		errQueue:      eq,
		in:            make(chan Message),
		out:           make(chan ProcessedMessage),
		bad:           make(chan BadMessage),
		control:       make(chan byte),
		maxGoroutines: mg,
	}
}

//Waits for new messages by BRPOP
func (c *Consumer) Wait4Messages() {
	for i := 1; i <= c.maxGoroutines; i++ {
		go c.RunWorker(i)
	}
	log.Printf("%d workers started.\n", c.maxGoroutines)

	pc := c.pool.Get()
	defer pc.Close()

	for {
		select {
		case s := <-c.control:
			switch s {
			case CTRL_STOP:
				close(c.in)
				close(c.bad)
				log.Println("All workers stopped.")
				return
			default:
				//Do nothing
			}
		default:
			msg, err := pc.Do("BRPOP", c.queue, 0)
			if err != nil {
				log.Println("unable to get message from redis: ", err)
			}

			v, err := redis.Values(msg, err)
			PanicIf(err)

			s, err := redis.String(v[1], err)
			PanicIf(err)

			c.in <- Message(s)
		}
	}
}

//Makes primary work for random milliseconds.
func (c *Consumer) RunWorker(wid int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	processTime := time.Duration(r.Intn(MAX_RAND_PROCESS_TIME))

	for m := range c.in {
		if prob() {
			c.bad <- BadMessage{msg: m, err: fmt.Sprintf("Error code %d", processTime)}
		} else {
			time.Sleep(time.Millisecond * processTime)
			c.out <- ProcessedMessage{duration: processTime, msg: m, worker: wid}
		}
	}
}

//Stops all workers
func (c *Consumer) Stop() {
	c.control <- CTRL_STOP
}

//Get bad messages from chan and call PushError
func (c *Consumer) Wait4Errors() {
	pc := c.pool.Get()
	defer pc.Close()

	for {
		select {
		case b := <-c.bad:
			_, err := pc.Do("LPUSH", c.errQueue, b.String())
			PanicIf(err)
		}
	}
}
