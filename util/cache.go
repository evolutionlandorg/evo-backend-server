package util

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func SetCache(ctx context.Context, key string, value []byte, ttl int) (err error) {
	cacheKey := fmt.Sprintf("evo:%s", key)
	_, err = SubPoolWithContextDo(ctx)("setex", cacheKey, ttl, string(value))
	return err
}

func GetCache(ctx context.Context, key string) []byte {
	cacheKey := fmt.Sprintf("evo:%s", key)
	if cache, err := redis.String(SubPoolWithContextDo(ctx)("get", cacheKey)); err != nil {
		return nil
	} else {
		return []byte(cache)
	}
}

func SetMap(ctx context.Context, key string, field string, value interface{}) {
	cacheKey := fmt.Sprintf("evo:%s", key)
	_, _ = SubPoolWithContextDo(ctx)("HSET", cacheKey, field, value)
}

func OnceTask(ctx context.Context, key string, ttl int, f func()) {
	if IncrCache(ctx, key, ttl) > 1 {
		return
	}
	defer DelCache(ctx, key)
	f()
}

func GetIntMap(ctx context.Context, key string, field string) int64 {
	cacheKey := fmt.Sprintf("evo:%s", key)
	n, _ := redis.Int64(SubPoolWithContextDo(ctx)("HGET", cacheKey, field))
	return n
}

func IncrCache(ctx context.Context, key string, ttl int) int {
	cacheKey := fmt.Sprintf("evo:%s", key)
	do := SubPoolWithContextDo(ctx)
	n, _ := redis.Int(do("Incr", cacheKey))
	if ttl > 0 {
		_, _ = do("EXPIRE", cacheKey, ttl)
	}
	return n
}

func DelCache(ctx context.Context, key string) {
	cacheKey := fmt.Sprintf("evo:%s", key)
	_, _ = SubPoolWithContextDo(ctx)("del", cacheKey)
}

func SaddCache(ctx context.Context, key, value string) bool {
	cacheKey := fmt.Sprintf("evo:%s", key)
	if intReturn, err := redis.Int(SubPoolWithContextDo(ctx)("sadd", cacheKey, value)); err != nil || intReturn != 1 {
		return false
	} else {
		return true
	}
}

func SremCache(ctx context.Context, key, value string) bool {
	cacheKey := fmt.Sprintf("evo:%s", key)
	if intReturn, err := redis.Int(SubPoolWithContextDo(ctx)("srem", cacheKey, value)); err != nil || intReturn != 1 {
		return false
	} else {
		return true
	}
}

func SmembersCache(ctx context.Context, key string) []string {
	cacheKey := fmt.Sprintf("evo:%s", key)
	intReturn, _ := redis.Strings(SubPoolWithContextDo(ctx)("smembers", cacheKey))
	return intReturn

}

func SaddArray(ctx context.Context, key string, value []interface{}) bool {
	data := []interface{}{fmt.Sprintf("evo:%s", key)}
	value = append(data, value...)
	if intReturn, err := redis.Int(SubPoolWithContextDo(ctx)("sadd", value...)); err != nil || intReturn < 1 {
		return false
	} else {
		return true
	}

}

func SRemArray(ctx context.Context, key string, value []interface{}) bool {
	data := []interface{}{fmt.Sprintf("evo:%s", key)}
	value = append(data, value...)
	if intReturn, err := redis.Int(SubPoolWithContextDo(ctx)("srem", value...)); err != nil || intReturn < 1 {
		return false
	} else {
		return true
	}

}

//func HgetCache(key, field string) []byte {
//	conn := subPool.Get()
//	defer conn.Close()
//	cacheKey := fmt.Sprintf("evo:%s", key)
//	if cache, err := redis.String(conn.Do("hget", cacheKey, field)); err != nil {
//		return nil
//	} else {
//		return []byte(cache)
//	}
//}

//func HsetCache(key, field string, value []byte) []byte {
//	conn := subPool.Get()
//	defer conn.Close()
//	cacheKey := fmt.Sprintf("evo:%s", key)
//	if cache, err := redis.String(conn.Do("hset", cacheKey, field, string(value))); err != nil {
//		return nil
//	} else {
//		return []byte(cache)
//	}
//}

//func HGetAllInt(key string) (ms map[string]int) {
//	conn := subPool.Get()
//	defer conn.Close()
//	ms, _ = redis.IntMap(conn.Do("HGETALL", key))
//	return
//}

//func HmSet(key string, value interface{}) (err error) {
//	conn := subPool.Get()
//	defer conn.Close()
//	args := redis.Args{}.Add(key)
//	switch v := value.(type) {
//	case map[string]string:
//		for k, v := range v {
//			args = args.Add(k).Add(v)
//		}
//	case map[string]int:
//		for k, v := range v {
//			args = args.Add(k).Add(v)
//		}
//	}
//	if len(args) <= 1 {
//		return
//	}
//	_, err = conn.Do("HMSET", args...)
//	return
//}
