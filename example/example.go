package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/orientlu/redis_rate"
	"time"
)

var (
	pool                  *redis.Pool
	redisDialWriteTimeout = time.Second
	redisDialReadTimeout  = time.Minute
)

func init() {
	pool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: time.Duration(300) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL("redis://:redis_@@@localhost:6379",
				redis.DialReadTimeout(redisDialReadTimeout),
				redis.DialWriteTimeout(redisDialWriteTimeout),
			)
		},
	}
}

func main() {
	limiter := redis_rate.NewLimiter(pool)
	go func() {
		for {
			res, err := limiter.Allow("project:123", redis_rate.PerSecond(100))
			if err != nil {
				panic(err)
			}
			fmt.Printf("%+v\n", res)
			time.Sleep(time.Millisecond * 15)
		}
	}()

	for {
		res, err := limiter.Allow("project:123", redis_rate.PerSecond(100))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", res)
		time.Sleep(time.Millisecond * 20)
	}

}
