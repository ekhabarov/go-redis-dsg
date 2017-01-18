## Distributed message consumer and generator for Redis.

### Dependencies

```
  go get github.com/garyburd/redigo/redis
  go get github.com/satori/go.uuid 
  go get github.com/ekhabarov/bbeye
```

### Install
```
  go get github.com/ekhabarov/go-redis-dsg
```

### Run
```
  . ./env; go-redis-dsg 
```
or 
```
MPORT=8082 MODE=generator go-dgen
```

#### Command line flags
* `-getErrors` - get all errors from `REDIS_ERROR_QUEUE` list, print them and 
delete them from Redis.

### Environment variables
Name | Type | Default | Description
-----|------|---------|------------
MPORT | integer | 8080 | Port for [expvarmon](https://github.com/divan/expvarmon) monitoring tool.
GENERATOR_PING_INTERVAL | integer | 10 | Number of seconds between ping requests while checking is generator working.
GENERATOR_INTERVAL | integer | 500 | Number of milliseconds between new messages.
REDIS_URL | string | '127.0.0.1:6379' | Redis address.
REDIS_QUEUE | string | 'a' | Redis list name which contains generated messages.
REDIS_ERROR_QUEUE | string | 'errors' | Redis list name with processes but invalid messages. 
REDIS_POOL_SIZE | int | 100 | Maximum number of connections to Redis
CONSUMER_MAX_GOROUTINES | integer | 10000 | Maximum number of goroutines using for processing messages.

### How it works
Application has two working modes: `generator` and `consumer`. 

#### Generator mode 
In this mode app generates messages and send it to Redis list with name 
provided by `REDIS_QUEUE` variable,  with `LPUSH` command, every 
`GENERATOR_INTERVAL` milliseconds. Generated message is just string of format 
`uuid + " " + timstamp`.

Only one generator can work at moment.

Generator could be stopped without any preparation.

#### Consumer mode
In this mode:

1. App waits messages from `REDIS_QUEUE` list using `BRPOP` command.
1. It processes  message within random number of milliseconds.
1. It sends processed messages into `out` Go chan and "invalid" messages to 
`bad` Go chan. From `bad` chan all messages transfer to `REDIS_ERROR_QUEUE` list.
1. Every `GENERATOR_PING_INTERVAL` seconds consumer sends ping requests to Redis,
to check if generator is working or not. If generator is working consumer will 
contoinue to work, if generator is not working, consumer will stop all workers 
and it will switch to generator mode.

In the end only one generator will working.


