package main

import (
	"log"
	"os"
	"strconv"
)

const (
	MODE                    = "MODE"
	GENERATOR_NAME          = "GENERATOR_NAME"
	GENERATOR_INTERVAL      = "GENERATOR_INTERVAL"
	GENERATOR_PING_INTERVAL = "GENERATOR_PING_INTERVAL"
	MULTIGEN                = "MULTIGEN"
	REDIS_URL               = "REDIS_URL"
	REDIS_QUEUE             = "REDIS_QUEUE"
	CONSUMER_MAX_GOROUTINES = "CONSUMER_MAX_GOROUTINES"
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
	cfg := &Config{
		redis:     &cfgRedis{},
		generator: &cfgGenerator{},
		consumer:  &cfgConsumer{},
	}

	cfg.mode = os.Getenv(MODE)

	cfg.redis.url = os.Getenv(REDIS_URL)
	cfg.redis.queue = os.Getenv(REDIS_QUEUE)

	cfg.generator.name = os.Getenv(GENERATOR_NAME)
	cfg.generator.multi = os.Getenv(MULTIGEN) != ""

	cfg.generator.interval = 500 //default value
	if envGI := os.Getenv(GENERATOR_INTERVAL); envGI != "" {
		if gi, err := strconv.Atoi(os.Getenv(GENERATOR_INTERVAL)); err != nil {
			log.Printf("invalid %s value: %s: using default", GENERATOR_INTERVAL, err)
		} else {
			cfg.generator.interval = gi
		}
	}

	cfg.generator.pingInterval = 10 //default value
	if envGPI := os.Getenv(GENERATOR_PING_INTERVAL); envGPI != "" {
		if gpi, err := strconv.Atoi(os.Getenv(GENERATOR_PING_INTERVAL)); err != nil {
			log.Printf("invalid %s value: %s: using default", GENERATOR_PING_INTERVAL, err)
		} else {
			cfg.generator.pingInterval = gpi
		}
	}

	cfg.consumer.maxGoroutines = 1000 //default value
	if envCMG := os.Getenv(CONSUMER_MAX_GOROUTINES); envCMG != "" {
		if maxGr, err := strconv.Atoi(envCMG); err != nil {
			log.Printf("invalid %s value: %s: using default", CONSUMER_MAX_GOROUTINES, err)
		} else {
			cfg.consumer.maxGoroutines = maxGr
		}
	}

	return cfg
}
