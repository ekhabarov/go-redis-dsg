package main

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

//Creates new redis.Pool
func NewRedisPool(addr string, rp int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
		MaxActive:   rp,
		Wait:        true,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Second*30 {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
