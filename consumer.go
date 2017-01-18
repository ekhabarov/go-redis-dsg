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

	MODE_UNKNOWN = iota
	MODE_GENERATOR
	MODE_CONSUMER
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
		stop          chan struct{}
		isActive      bool
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
		maxGoroutines: mg,
		stop:          make(chan struct{}),
	}
}

func (c *Consumer) Process(in chan Message, out chan ProcessedMessage) {
	log.Println("Preparing for starting consumer")

	c.in = in
	c.out = out

	//TODO Add PING before run workers
	for i := 1; i <= c.maxGoroutines; i++ {
		go c.RunWorker(i)
	}
	log.Printf("%d workers started.\n", c.maxGoroutines)

	c.isActive = true

	pc := c.pool.Get()
	defer pc.Close()
	defer func(c *Consumer) {
		log.Println("Consumer stopped.")
		c.isActive = false
	}(c)
	defer c.Close()

	for {
		select {
		case <-c.stop:
			log.Println("Consumer stopping")
			time.Sleep(time.Second)
			c.stop <- struct{}{}
			return
		default:
			msg, err := pc.Do("BRPOP", c.queue, 0)
			if err != nil {
				log.Println("messages: unable to get from redis:", err)
				return
			}

			v, err := redis.Values(msg, err)
			LogIf(err)

			s, err := redis.String(v[1], err)
			LogIf(err)

			fmt.Printf(":")
			c.in <- Message(s)
			fmt.Printf(".")
		}
	}
}

//Waits for new messages by BRPOP
func (c *Consumer) Start() chan ProcessedMessage {
	in := make(chan Message)
	out := make(chan ProcessedMessage)
	bad := make(chan BadMessage)
	go c.ProcessErrors(bad)
	go c.Process(in, out)
	return out
}

//Stops all workers
func (c *Consumer) Stop() {
	if !c.IsActive() {
		return
	}
	log.Println("stop: 1")
	c.stop <- struct{}{}
	log.Println("stop: 2")
	<-c.stop
	log.Println("stop: 3")
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
			fmt.Printf(">")
			c.out <- ProcessedMessage{duration: processTime, msg: m, worker: wid}
			fmt.Printf("<")
		}
	}
}

func (c *Consumer) Close() {
	if !c.IsActive() {
		return
	}
	close(c.bad)
	close(c.in)
	close(c.out)
	log.Println("Channels closed.")
}

//Get bad messages from chan and call PushError
func (c *Consumer) ProcessErrors(bad chan BadMessage) {
	pc := c.pool.Get()
	defer pc.Close()

	c.bad = bad

	for b := range c.bad {
		_, err := pc.Do("LPUSH", c.errQueue, b.String())
		if err != nil {
			log.Println("errors: unable to push to redis:", err)
		}
	}
}

func (c *Consumer) IsActive() bool {
	return c.isActive
}
