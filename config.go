package main

import (
	"log"
	"os"
	"strconv"
)

const (
	GENERATOR_INTERVAL      = "GENERATOR_INTERVAL"
	GENERATOR_PING_INTERVAL = "GENERATOR_PING_INTERVAL"
	REDIS_URL               = "REDIS_URL"
	REDIS_QUEUE             = "REDIS_QUEUE"
	REDIS_ERROR_QUEUE       = "REDIS_ERROR_QUEUE"
	REDIS_POOL_SIZE         = "REDIS_POOL_SIZE"
	CONSUMER_MAX_GOROUTINES = "CONSUMER_MAX_GOROUTINES"
)

type (
	cfgRedis struct {
		url      string
		queue    string
		errQueue string
		poolSize int
	}

	cfgGenerator struct {
		interval     int
		pingInterval int
	}

	cfgConsumer struct {
		maxGoroutines int
	}

	Config struct {
		redis     *cfgRedis
		generator *cfgGenerator
		consumer  *cfgConsumer
	}
)

func readIntParam(val *int, def int, e string) {
	*val = def //default value
	if v := os.Getenv(e); v != "" {
		if iv, err := strconv.Atoi(v); err != nil {
			log.Printf("invalid %s value: %s: using default", e, err)
		} else {
			*val = iv
		}
	}
}

//Read environmetns variables and returns new Config struct
func ReadConfig() *Config {
	cfg := &Config{
		redis:     &cfgRedis{},
		generator: &cfgGenerator{},
		consumer:  &cfgConsumer{},
	}

	cfg.redis.url = os.Getenv(REDIS_URL)
	cfg.redis.queue = os.Getenv(REDIS_QUEUE)
	cfg.redis.errQueue = os.Getenv(REDIS_ERROR_QUEUE)

	readIntParam(&cfg.redis.poolSize, 100, REDIS_POOL_SIZE)
	readIntParam(&cfg.generator.interval, 500, GENERATOR_INTERVAL)
	readIntParam(&cfg.generator.pingInterval, 10, GENERATOR_PING_INTERVAL)
	readIntParam(&cfg.consumer.maxGoroutines, 1000, CONSUMER_MAX_GOROUTINES)

	return cfg
}
