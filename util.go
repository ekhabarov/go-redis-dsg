package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

//Call panic if err is not nil
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

//Prints error to stdout if err is not nil
func LogIf(err error, v ...string) {
	if err != nil {
		log.Println(strings.Join(v, ":"), err)
	}
}

//Emulates 5% probability
func prob() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(100)
	return n <= 5
}

func Mode(g *Generator, c *Consumer) byte {
	if g.IsActive() {
		return MODE_GENERATOR
	}
	if c.IsActive() {
		return MODE_CONSUMER
	} else {
		return MODE_UNKNOWN
	}

}

//Pinger
func StartPing(g *Generator, c *Consumer, p *chan ProcessedMessage) {
	ticker := time.NewTicker(time.Second * time.Duration(g.pingInterval))

	for range ticker.C {

		m := Mode(g, c)
		fmt.Println("ping: ", m, "c.IsActive:", c.IsActive(), "g.IsActive:", g.IsActive())

		switch m {
		case MODE_GENERATOR:
			if !g.AcquireLock() {
				switch g.RefreshLock() {
				case LOCK_NOT_REFRESHED:
					g.Stop()
					fmt.Printf("Switching to consumer")
					*p = c.Start()
				}
			}
		case MODE_CONSUMER:
			if g.AcquireLock() {
				fmt.Printf("Switching to generator")
				c.Stop()
				go g.Start()
			}
		default:
			if g.AcquireLock() {
				go g.Start()
			} else {
				*p = c.Start()
			}

			//panic("unreachable mode")
		}

		//can be gen->cons->gen->cons transfer? yes

		//if mode = consumer try to acquire lock
		//	if true stop consumer, start generator
		//	if false start consumer

		//if mode = generator try to acquire lock,
		//	if false try to refresh lock
		//	if false stop generator, run consumer

	}
}
