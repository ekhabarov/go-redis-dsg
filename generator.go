package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

type Generator struct {
	pool     *redis.Pool
	queue    string
	name     string
	interval int
}

func NewGen(p *redis.Pool, queue string, name string, interval int) *Generator {
	return &Generator{
		pool:     p,
		queue:    queue,
		name:     name,
		interval: interval,
	}
}

func (g *Generator) Exists() bool {
	pc := g.pool.Get()
	defer pc.Close()

	data, err := pc.Do("CLIENT", "LIST")
	PanicIf(err)

	v, err := redis.Bytes(data, err)
	PanicIf(err)

	s, err := redis.String(v, err)
	PanicIf(err)

	return strings.Contains(s, "name="+g.name)
}

func (g *Generator) Connect(multi bool) error {
	if !multi && g.Exists() {
		return errors.New("Another generator process already in progress. Exiting...")
	}

	pc := g.pool.Get()
	defer pc.Close()

	_, err := pc.Do("CLIENT", "SETNAME", g.name)
	PanicIf(err)

	ticker := time.NewTicker(time.Millisecond * time.Duration(g.interval))
	go g.Run(ticker.C)

	return nil
}

//Generates random string
func (g *Generator) Message() string {
	return fmt.Sprintf("%s at %s", uuid.NewV4(), time.Now())
}

//Added data to redis list
func (g *Generator) Run(t <-chan time.Time) {
	pc := g.pool.Get()
	defer pc.Close()
	for range t {

		pc.Do("LPUSH", g.queue, g.Message())
	}
}
