package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/ekhabarov/bbeye"
)

const (
	MODE_GENERATOR = "generator"
	MODE_CONSUMER  = "consumer"
)

func main() {
	runtime.GOMAXPROCS(1)

	cfg := ReadConfig()
	done := make(chan struct{})

	redisPool := NewRedisPool(cfg.redis.url)

	c := NewConsumer(redisPool, cfg.redis.queue, cfg.consumer.maxGoroutines)
	g := NewGen(redisPool, cfg.redis.queue, cfg.generator.name, cfg.generator.interval)

	switch cfg.mode {

	case MODE_GENERATOR:
		if err := g.Connect(cfg.generator.multi); err != nil {
			log.Fatalln(err)
		}

	case MODE_CONSUMER:
		go bbeye.Run("127.0.0.1:" + os.Getenv("MPORT"))
		go c.Wait4Messages()

		//Ping generator
		go func(g *Generator, c *Consumer) {
			for {
				select {
				case <-time.After(time.Second * time.Duration(cfg.generator.pingInterval)):
					if !g.Exists() {
						if err := g.Connect(cfg.generator.multi); err != nil {
							log.Fatalln("unreachable error:", err)
						}
						c.Stop()
						return
					}
				}
			}
		}(g, c)

	default:
		log.Fatalln("invalid mode: ", cfg.mode)
	}

	//Exit timeout
	go func(d chan<- struct{}) {
		time.Sleep(time.Second * 180)
		d <- struct{}{}
	}(done)

	for {
		select {
		case <-c.out:
			//Process finished task

		case <-done:
			fmt.Println("Done")
			return
		}
	}
}
