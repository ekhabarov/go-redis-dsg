package main

import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

const (
	//Key lock name.
	LOCK_NAME = "lock:gen"

	//Another generator is active
	LOCK_OTHER_EXISTS = iota

	//Lock have been successfully refreshed
	LOCK_REFRESHED

	//Lock exists, but belongs to another generator
	LOCK_NOT_REFRESHED
)

type Generator struct {
	pool         *redis.Pool
	queue        string
	name         string
	interval     int
	pingInterval int
	stop         chan struct{}
	isActive     bool
}

//Creates new Generator
func NewGen(p *redis.Pool, queue string, interval, pinterval int) *Generator {
	return &Generator{
		pool:         p,
		queue:        queue,
		interval:     interval,
		pingInterval: pinterval,
		name:         uuid.NewV4().String(),
		stop:         make(chan struct{}),
	}
}

//Tries to acquire new lock. If there is no lock, that means there is no
//active generator. Acquiring lock means cuurrent generator will be activated.
func (g *Generator) AcquireLock() bool {
	pc := g.pool.Get()
	defer pc.Close()

	//Set lock expire time pingInterval+5 seconds.
	//5 is a "magic number", time for network lag, etc.
	lock, err := pc.Do("SET", LOCK_NAME, g.name, "EX", g.pingInterval+5, "NX")
	if err != nil {
		log.Println("acquirelock: unable to execute SET:", err)
		return false
	}

	return lock == "OK"
}

//Refreshes it's own lock
func (g *Generator) RefreshLock() byte {
	pc := g.pool.Get()
	defer pc.Close()

	_, err := pc.Do("WATCH", LOCK_NAME)
	if err != nil {
		log.Println("refreshlock: unable to execute WATCH:", err)
		return LOCK_NOT_REFRESHED
	}

	cg, err := pc.Do("GET", LOCK_NAME)
	if err != nil {
		log.Println("refreshlock: unable to execute GET:", err)
		return LOCK_NOT_REFRESHED
	}

	cg, err = redis.String(cg, err)
	LogIf(err)

	if cg == g.name {
		pc.Send("MULTI")
		pc.Send("SET", LOCK_NAME, g.name, "EX", g.pingInterval+5, "XX")
		lock, err := redis.Values(pc.Do("EXEC"))
		if err != nil {
			log.Println("refreshlock: unable to execute EXEC:", err)
			return LOCK_NOT_REFRESHED
		}

		var s string
		redis.Scan(lock, &s)

		if s == "OK" {
			return LOCK_REFRESHED
		} else {
			return LOCK_NOT_REFRESHED
		}

	} else {
		_, err := pc.Do("UNWATCH")
		if err != nil {
			log.Println("refreshlock: unable to execute UNWATCH:", err)
			return LOCK_OTHER_EXISTS
		}
	}
	return LOCK_NOT_REFRESHED
}

//Generates random string
func (g *Generator) Message() string {
	return fmt.Sprintf("%s at %s", uuid.NewV4(), time.Now())
}

//Added data to redis list
func (g *Generator) Start() {
	if g.RefreshLock() == LOCK_OTHER_EXISTS {
		log.Println("generator run: another generator in progress")
		return
	}

	pc := g.pool.Get()
	defer pc.Close()

	defer func(g *Generator) {
		log.Printf("Generator stopped (name: %s).\n", g.name)
		g.isActive = false
	}(g)

	fmt.Println("Interval:", g.interval)
	ticker := time.NewTicker(time.Millisecond * time.Duration(g.interval))

	log.Printf("Generator started (name: %s).\n", g.name)
	g.isActive = true
	i := 0
	for {
		select {
		case <-ticker.C:
			i++
			if i >= 1000 {
				fmt.Printf("+")
				i = 0
			}
			fmt.Printf(".")
			if _, err := pc.Do("LPUSH", g.queue, g.Message()); err != nil {
				log.Println("messages: unable to push to redis:", err)
				return
			}
		case <-g.stop:
			return
		}
	}
}

func (g *Generator) Stop() {
	g.stop <- struct{}{}
}

func (g *Generator) IsActive() bool {
	return g.isActive
}
