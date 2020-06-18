package redis_rate

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

const redisPrefix = "rate:"

// rediser redisgo conn pool
type rediser interface {
	Get() redis.Conn
}

// Limit set Limit
type Limit struct {
	Rate   int
	Period time.Duration
	Burst  int
}

// PerSecond Limit for second
func PerSecond(rate int) *Limit {
	return &Limit{
		Rate:   rate,
		Period: time.Second,
		Burst:  rate,
	}
}

// PerMinute Limit for Minute
func PerMinute(rate int) *Limit {
	return &Limit{
		Rate:   rate,
		Period: time.Minute,
		Burst:  rate,
	}
}

// PerHour Limit for Hour
func PerHour(rate int) *Limit {
	return &Limit{
		Rate:   rate,
		Period: time.Hour,
		Burst:  rate,
	}
}

//------------------------------------------------------------------------------

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	rdb rediser
}

// NewLimiter returns a new Limiter.
func NewLimiter(rdb rediser) *Limiter {
	return &Limiter{
		rdb: rdb,
	}
}

// Allow is a shortcut for AllowN(ctx, key, limit, 1).
func (l *Limiter) Allow(key string, limit *Limit) (*Result, error) {
	return l.AllowN(key, limit, 1)
}

// AllowN reports whether n events may happen at time now.
func (l *Limiter) AllowN(
	key string,
	limit *Limit,
	n int,
) (*Result, error) {
	c := l.rdb.Get()
	defer c.Close()

	values := []interface{}{string(redisPrefix + key), limit.Burst, limit.Rate, limit.Period.Seconds(), n}
	v, err := allowN.Do(c, values...)
	if err != nil {
		return nil, err
	}

	values = v.([]interface{})

	retryAfter, err := strconv.ParseFloat(string(values[2].([]byte)), 64)
	if err != nil {
		return nil, err
	}

	resetAfter, err := strconv.ParseFloat(string(values[3].([]byte)), 64)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Limit:      limit,
		Allowed:    int(values[0].(int64)),
		Remaining:  int(values[1].(int64)),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}
	return res, nil
}

// AllowAtMost reports whether at most n events may happen at time now.
// It returns number of allowed events that is less than or equal to n.
func (l *Limiter) AllowAtMost(
	key string,
	limit *Limit,
	n int,
) (*Result, error) {

	c := l.rdb.Get()
	defer c.Close()

	values := []interface{}{string(redisPrefix + key), limit.Burst, limit.Rate, limit.Period.Seconds(), n}
	v, err := allowAtMost.Do(c, values...)
	if err != nil {
		return nil, err
	}

	values = v.([]interface{})

	retryAfter, err := strconv.ParseFloat(string(values[2].([]byte)), 64)
	if err != nil {
		return nil, err
	}

	resetAfter, err := strconv.ParseFloat(string(values[3].([]byte)), 64)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Limit:      limit,
		Allowed:    int(values[0].(int64)),
		Remaining:  int(values[1].(int64)),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}
	return res, nil
}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}
	return time.Duration(f * float64(time.Second))
}

type Result struct {
	// Limit is the limit that was used to obtain this result.
	Limit *Limit

	// Allowed reports whether event may happen at time now.
	Allowed int

	// Remaining is the maximum number of requests that could be
	// permitted instantaneously for this key given the current
	// state. For example, if a rate limiter allows 10 requests per
	// second and has already received 6 requests for this key this
	// second, Remaining would be 4.
	Remaining int

	// RetryAfter is the time until the next request will be permitted.
	// It should be -1 unless the rate limit has been exceeded.
	RetryAfter time.Duration

	// ResetAfter is the time until the RateLimiter returns to its
	// initial state for a given key. For example, if a rate limiter
	// manages requests per second and received one request 200ms ago,
	// Reset would return 800ms. You can also think of this as the time
	// until Limit and Remaining will be equal.
	ResetAfter time.Duration
}
