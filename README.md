# Rate limiting for redisgo

This package is based on [rwz/redis-gcra](https://github.com/rwz/redis-gcra) and implements [GCRA](https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm) (aka leaky bucket) for rate limiting based on Redis. The code requires Redis version 3.2 or newer since it relies on [replicate_commands](https://redis.io/commands/eval#replicating-commands-instead-of-scripts) feature.

fork from [redis_rate for go-redis](https://github.com/go-redis/redis_rate)

## Installation

redis_rate requires a Go version with [Modules](https://github.com/golang/go/wiki/Modules) support and uses import versioning. So please make sure to initialize a Go module before installing redis_rate:

``` shell
go mod init github.com/my/repo
go get github.com/orientlu/redis_rate
```

Import:

``` go
import "github.com/orientlu/redis_rate"
```

## Example

``` go
import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/orientlu/redis_rate"
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

		res, err := limiter.Allow("project:123", redis_rate.PerSecond(100))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", res)
}
```
