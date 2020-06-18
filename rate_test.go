package redis_rate_test

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"

	"github.com/orientlu/redis_rate"
)

var (
	redisDialWriteTimeout = time.Second
	redisDialReadTimeout  = time.Minute
	redisURL              = "redis://:redis_@@@localhost:6379"
)

func rateLimiter() *redis_rate.Limiter {

	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: time.Duration(300) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL,
				redis.DialReadTimeout(redisDialReadTimeout),
				redis.DialWriteTimeout(redisDialWriteTimeout),
			)
		},
	}

	c := pool.Get()
	defer c.Close()
	if _, err := c.Do("del", "rate:test_id"); err != nil {
		panic(err)
	}
	return redis_rate.NewLimiter(pool)
}

func TestAllow(t *testing.T) {

	l := rateLimiter()
	limit := redis_rate.PerSecond(10)

	res, err := l.Allow("test_id", limit)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 1)
	assert.Equal(t, res.Remaining, 9)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 100*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 2)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 2)
	assert.Equal(t, res.Remaining, 7)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 300*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 7)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 7)
	assert.Equal(t, res.Remaining, 0)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 1000)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 0)
	assert.Equal(t, res.Remaining, 0)
	assert.InDelta(t, res.RetryAfter, 99*time.Second, float64(time.Second))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))
}

func TestAllowAtMost(t *testing.T) {

	l := rateLimiter()
	limit := redis_rate.PerSecond(10)

	res, err := l.Allow("test_id", limit)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 1)
	assert.Equal(t, res.Remaining, 9)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 100*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowAtMost("test_id", limit, 2)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 2)
	assert.Equal(t, res.Remaining, 7)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 300*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 0)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 0)
	assert.Equal(t, res.Remaining, 7)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 300*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowAtMost("test_id", limit, 10)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 7)
	assert.Equal(t, res.Remaining, 0)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 0)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 0)
	assert.Equal(t, res.Remaining, 0)
	assert.Equal(t, res.RetryAfter, time.Duration(-1))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowAtMost("test_id", limit, 1000)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 0)
	assert.Equal(t, res.Remaining, 0)
	assert.InDelta(t, res.RetryAfter, 99*time.Millisecond, float64(10*time.Millisecond))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))

	res, err = l.AllowN("test_id", limit, 1000)
	assert.Nil(t, err)
	assert.Equal(t, res.Allowed, 0)
	assert.Equal(t, res.Remaining, 0)
	assert.InDelta(t, res.RetryAfter, 99*time.Second, float64(time.Second))
	assert.InDelta(t, res.ResetAfter, 999*time.Millisecond, float64(10*time.Millisecond))
}

func BenchmarkAllow(b *testing.B) {
	l := rateLimiter()
	limit := redis_rate.PerSecond(1e6)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := l.Allow("foo", limit)
			if err != nil {
				b.Fatal(err)
			}
			if res.Allowed == 0 {
				panic("not reached")
			}
		}
	})
}

func BenchmarkAllowAtMost(b *testing.B) {
	l := rateLimiter()
	limit := redis_rate.PerSecond(1e6)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := l.AllowAtMost("foo", limit, 1)
			if err != nil {
				b.Fatal(err)
			}
			if res.Allowed == 0 {
				panic("not reached")
			}
		}
	})
}
