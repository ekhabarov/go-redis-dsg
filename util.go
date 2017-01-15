package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
)

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func MemPrint() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	fmt.Printf("Alloc: %d\t TotalAlloc: %d\t Head: %d\t HeapSys: %d\n", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
}

//Creates new redis.Pool
func NewRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

//Emulates 5% probability
func prob() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(100)
	return n <= 5
}

func pingGenerator(g *Generator, c *Consumer, interval int) {
	for {
		select {
		case <-time.After(time.Second * time.Duration(interval)):
			if !g.Exists() {
				if err := g.Connect(false); err != nil {
					log.Fatalln("unreachable error:", err)
				}
				c.Stop()
				log.Printf("Mode switched: consumer -> generator(name: %s).", g.name)
				return
			}
		}
	}
}
