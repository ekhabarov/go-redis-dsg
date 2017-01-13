package main

import (
	"log"
	"os"
	"strconv"
)

type (
	cfgRedis struct {
		url   string
		queue string
	}
	cfgGenerator struct {
		name         string
		multi        bool
		interval     int
		pingInterval int
	}
	cfgConsumer struct {
		maxGoroutines int
	}
	Config struct {
		mode      string
		redis     *cfgRedis
		generator *cfgGenerator
		consumer  *cfgConsumer
	}
)

func ReadConfig() *Config {
	var err error
	cfg := &Config{
		redis:     &cfgRedis{},
		generator: &cfgGenerator{},
		consumer:  &cfgConsumer{},
	}

	cfg.mode = os.Getenv("MODE")

	cfg.redis.url = os.Getenv("REDIS_URL")
	cfg.redis.queue = os.Getenv("REDIS_QUEUE")

	cfg.generator.name = os.Getenv("GENERATOR_NAME")
	cfg.generator.multi = os.Getenv("MULTIGEN") != ""

	cfg.generator.interval, err = strconv.Atoi(os.Getenv("GENERATOR_INTERVAL"))
	if err != nil {
		log.Println("invalid GENERATOR_INTERVAL value: ", err, ": using default")
		cfg.generator.interval = 500
	}

	cfg.consumer.maxGoroutines, err = strconv.Atoi(os.Getenv("MAX_GOROUTINES"))
	if err != nil {
		log.Println("invalid MAX_GOROUTINES value: ", err, ": using default")
		cfg.consumer.maxGoroutines = 1000
	}

	return cfg
}
