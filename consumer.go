package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	MAX_RAND_PROCESS_TIME = 1000

	CTRL_STOP = iota
)

type (
	Message string

	FinishedTask struct {
		duration time.Duration
		msg      Message
		worker   int
	}

	Consumer struct {
		pool          *redis.Pool
		queue         string
		in            chan Message
		out           chan FinishedTask
		maxGoroutines int
		control       chan byte
	}
)

func NewConsumer(p *redis.Pool, queue string, mg int) *Consumer {
	return &Consumer{
		pool:          p,
		queue:         queue,
		in:            make(chan Message),
		out:           make(chan FinishedTask),
		control:       make(chan byte),
		maxGoroutines: mg,
	}
}

//Waits for new messages by BRPOP
func (c *Consumer) Wait4Messages() {
	fmt.Println("Waiting for messages. Run workers.")
	for i := 1; i <= c.maxGoroutines; i++ {
		go c.RunWorker(i)
	}
	pc := c.pool.Get()
	defer pc.Close()

	for {
		select {
		case s := <-c.control:
			switch s {
			case CTRL_STOP:
				fmt.Println("Workers stoped")
				close(c.in)
				return
			default:
				//Do nothing
			}
		default:
			//fmt.Printf("+")
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
//Stop work by closing in channel
func (c *Consumer) RunWorker(wid int) {
	fmt.Printf(".")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	processTime := time.Duration(r.Intn(MAX_RAND_PROCESS_TIME))

	for t := range c.in {
		time.Sleep(time.Millisecond * processTime)
		c.out <- FinishedTask{processTime, t, wid}
	}
}

func (c *Consumer) Stop() {
	fmt.Println("Stop all workers")
	c.control <- CTRL_STOP
}
