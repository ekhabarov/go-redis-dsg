package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
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
	redisQueue := os.Getenv("REDIS_QUEUE")
	done := make(chan struct{})
	taskChan := make(chan Message)
	taskDone := make(chan finishedTask)

	interval, err := strconv.Atoi(os.Getenv("INTERVAL_MS"))
	if err != nil {
		log.Println("invalid interval value: ", err, ": using default")
		interval = 500
	}

	redisConn, err := redis.DialURL(os.Getenv("REDIS_URL"))
	PanicIf(err)
	defer redisConn.Close()

	mode := os.Getenv("MODE")
	switch mode {

	//RUN GENERATOR
	case MODE_GENERATOR:
		name := os.Getenv("GENERATOR_NAME")
		multi := os.Getenv("MULTIGEN") != ""

		g := NewGen(redisConn, redisQueue, name, interval)
		if err := g.Connect(multi); err != nil {
			log.Fatalln(err)
		}

	//RUN CONSUMER
	case MODE_CONSUMER:
		maxGoroutines, err := strconv.Atoi(os.Getenv("MAX_GOROUTINES"))
		if err != nil {
			log.Println("invalid MAX_GOROUTINES value: ", err, ": using default")
			maxGoroutines = 1000
		}

		for i := 0; i <= maxGoroutines; i++ {
			go worker(i, taskChan, taskDone)
		}

		go waiter(redisConn, redisQueue, taskChan)
		//go bbeye.Run("127.0.0.1:8080")
	default:
		log.Fatalln("invalid mode: ", mode)
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
