package main

import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

const (
	LOCK_NAME = "lock:gen"
)

type Generator struct {
	pool         *redis.Pool
	queue        string
	name         string
	interval     int
	pingInterval int
}

//Creates new Generator
func NewGen(p *redis.Pool, queue string, interval, pinterval int) *Generator {
	return &Generator{
		pool:         p,
		queue:        queue,
		interval:     interval,
		pingInterval: pinterval,
		name:         uuid.NewV4().String(),
	}
}

//Tries to acquire new lock
func (g *Generator) AcquireLock() bool {
	pc := g.pool.Get()
	defer pc.Close()

	//Set lock expire time pingInterval+5 seconds.
	//5 is a "magic number", time for network lag, etc.
	lock, err := pc.Do("SET", LOCK_NAME, g.name, "EX", g.pingInterval+5, "NX")
	LogIf(err)

	return lock == "OK"
}

//Refreshes it's own lock
func (g *Generator) RefreshLock() bool {
	pc := g.pool.Get()
	defer pc.Close()

	_, err := pc.Do("WATCH", LOCK_NAME)
	PanicIf(err)

	cg, err := pc.Do("GET", LOCK_NAME)
	PanicIf(err)

	cg, err = redis.String(cg, err)
	PanicIf(err)

	if cg == g.name {
		pc.Send("MULTI")
		pc.Send("SET", LOCK_NAME, g.name, "EX", g.pingInterval+5, "XX")
		lock, err := pc.Do("EXEC")
		PanicIf(err)

		return lock == "OK"
	} else {
		_, err := pc.Do("UNWATCH")
		PanicIf(err)
	}
	return false
}

//Generates random string
func (g *Generator) Message() string {
	return fmt.Sprintf("%s at %s", uuid.NewV4(), time.Now())
}

//Added data to redis list
func (g *Generator) Run() {
	pc := g.pool.Get()
	defer pc.Close()

	ticker := time.NewTicker(time.Millisecond * time.Duration(g.interval))

	log.Printf("Generator started (name: %s).\n", g.name)

	for range ticker.C {
		pc.Do("LPUSH", g.queue, g.Message())
	}
}

//Pinger
func (g *Generator) Pinger(c *Consumer) {
	for {
		select {
		case <-time.After(time.Second * time.Duration(g.pingInterval)):
			if g.AcquireLock() {
				go g.Run()
				c.Stop()
			} else {
				g.RefreshLock()
			}
		}
	}
}
