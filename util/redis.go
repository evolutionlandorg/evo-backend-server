package util

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/itering/go-workers"
	"golang.org/x/net/context"

	"github.com/go-redsync/redsync/v4"
	"github.com/gomodule/redigo/redis"
	redigotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gomodule/redigo"
)

var (
	subPool *redis.Pool
	RedSync *redsync.Redsync

	redisHost     = GetEnv("REDIS_HOST", "127.0.0.1")
	redisPort     = GetEnv("REDIS_PORT", "6379")
	redisPassword = GetEnv("REDIS_PASSWORD", "")
	redisDatabase = GetEnv("REDIS_DATABASE", "0")
)

func init() {

}

func InitWorkers() error {
	workers.Configure(map[string]string{
		"server":    redisHost + ":" + redisPort,
		"database":  redisDatabase,
		"pool":      "30",
		"process":   "1",
		"namespace": "evo",
	})
	return nil
}

func InitRedis() error {
	db, _ := strconv.Atoi(redisDatabase)
	subPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redigotrace.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort),
				redis.DialPassword(redisPassword), redis.DialDatabase(db))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	c := subPool.Get()
	defer c.Close()
	RedSync = redsync.New(redigo.NewPool(subPool))
	_, err := c.Do("ping")
	return err
}

func SubPoolWithContextDo(ctx context.Context) func(commandName string, args ...interface{}) (reply interface{}, err error) {
	return func(commandName string, args ...interface{}) (reply interface{}, err error) {
		conn := subPool.Get()
		defer conn.Close()
		args = append(args, ctx)
		return conn.Do(commandName, args...)
	}
}
