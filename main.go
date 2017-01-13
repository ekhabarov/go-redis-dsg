package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	MODE_GENERATOR = "generator"
	MODE_CONSUMER  = "consumer"
)

type finishedTask struct {
	duration time.Duration
	msg      Message
	worker   int
}

func main() {
	runtime.GOMAXPROCS(1)

	//Redis queue which contains tasks
	cfg := ReadConfig()
	done := make(chan struct{})
	taskChan := make(chan Message)
	taskDone := make(chan finishedTask)

	redisConn, err := redis.DialURL(cfg.redis.url)
	PanicIf(err)
	defer redisConn.Close()

	switch cfg.mode {

	//RUN GENERATOR
	case MODE_GENERATOR:

		g := NewGen(redisConn, cfg.redis.queue, cfg.generator.name, cfg.generator.interval)
		if err := g.Connect(cfg.generator.multi); err != nil {
			log.Fatalln(err)
		}

	//RUN CONSUMER
	case MODE_CONSUMER:

		for i := 0; i <= cfg.consumer.maxGoroutines; i++ {
			go worker(i, taskChan, taskDone)
		}

		go waiter(redisConn, cfg.redis.queue, taskChan)
		//go bbeye.Run("127.0.0.1:8080")
	default:
		log.Fatalln("invalid mode: ", cfg.mode)
	}

	//MemPrint()
	go func() {
		time.Sleep(time.Second * 90)
		done <- struct{}{}
	}()

	for {
		select {
		case <-taskDone:
			//fmt.Printf("Done for %dms by worker %d.\n", f.duration, f.worker)
		case <-time.After(time.Second * 10):
			fmt.Println("Ping generator")
		//MemPrint()
		case <-done:
			fmt.Println("Done")
			return
		}

	}
}
