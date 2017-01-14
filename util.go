package main

import (
	"fmt"
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

func NewRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}
