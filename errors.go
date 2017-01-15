package main

import "github.com/garyburd/redigo/redis"

type (
	ErrorReader struct {
		pool     *redis.Pool
		errQueue string
	}

	EMsg struct {
		M string
		E string
	}
)

//Creates new ErrorReader
func NewErrorReader(p *redis.Pool, eq string) *ErrorReader {
	return &ErrorReader{
		pool:     p,
		errQueue: eq,
	}
}

//Returns number of errors
func (r *ErrorReader) Count() int64 {
	pc := r.pool.Get()
	defer pc.Close()

	d, err := pc.Do("LLEN", r.errQueue)
	LogIf(err)

	q, err := redis.Int64(d, err)
	LogIf(err)

	return q
}

//Read all errors and clear list
func (r *ErrorReader) ReadErrors() (list []string) {
	pc := r.pool.Get()
	defer pc.Close()

	pc.Send("MULTI")
	pc.Send("LRANGE", r.errQueue, "0", "-1")
	pc.Send("DEL", r.errQueue)

	v, err := pc.Do("EXEC")
	PanicIf(err)

	_, err = redis.Scan(v.([]interface{}), &list)
	PanicIf(err)

	return list
}
