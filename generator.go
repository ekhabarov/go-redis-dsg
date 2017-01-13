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
	conn     redis.Conn
	queue    string
	name     string
	interval int
}

func NewGen(c redis.Conn, queue string, name string, interval int) *Generator {
	return &Generator{
		conn:     c,
		queue:    queue,
		name:     name,
		interval: interval,
	}
}

func (g *Generator) Exists() bool {
	data, err := g.conn.Do("CLIENT", "LIST")
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

	_, err := g.conn.Do("CLIENT", "SETNAME", g.name)
	PanicIf(err)
	fmt.Println("Generator started. Using name: ", g.name)

	ticker := time.NewTicker( /*time.Millisecond * */ time.Duration(g.interval))

	go g.Run(ticker.C)

	return nil
}

//Generates random string
func (g *Generator) Message() string {
	return fmt.Sprintf("%s at %s", uuid.NewV4(), time.Now())
}

//Added data to redis list
func (g *Generator) Run(t <-chan time.Time) {
	for range t {
		g.conn.Do("LPUSH", g.queue, g.Message())
	}
}
