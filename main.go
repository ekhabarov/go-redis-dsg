//Distributed message consumer and generator for Redis.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/ekhabarov/bbeye"
)

// TODO:
// 1. Rebuild generator determination mechanism
// 2. One connection to Redis on error
// 3. Remove MODE param. Determine role by generataor existing
const (
	MODE_GENERATOR = "generator"
	MODE_CONSUMER  = "consumer"
)

func main() {
	runtime.GOMAXPROCS(1)
	//runtime.GOMAXPROCS(runtime.NumCPU())

	ge := flag.Bool("getErrors", false, "Get all errors from Redis")
	flag.Parse()

	cfg := ReadConfig()
	done := make(chan struct{})

	redisPool := NewRedisPool(cfg.redis.url)

	if *ge {
		e := NewErrorReader(redisPool, cfg.redis.errQueue)
		if e.Count() < 1 {
			log.Println("Errors list is empty.")
			return
		}
		for _, m := range e.ReadErrors() {
			fmt.Println(m)
		}
		return
	}

	c := NewConsumer(redisPool, cfg.redis.queue, cfg.redis.errQueue, cfg.consumer.maxGoroutines)
	g := NewGen(redisPool, cfg.redis.queue, cfg.generator.interval, cfg.generator.pingInterval)

	if g.AcquireLock() {
		go g.Run()
	} else {
		go bbeye.Run("127.0.0.1:" + os.Getenv("MPORT"))
		go c.Wait4Errors()
		go c.Wait4Messages()
	}
	go g.Pinger(c)

	//Exit timeout
	go func(d chan<- struct{}) {
		time.Sleep(time.Second * 300)
		d <- struct{}{}
	}(done)

	for {
		select {
		case <-c.out:
		case <-done:
			log.Println("Done.")
			return
		}
	}
}
