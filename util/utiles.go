package util

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"time"
)

// If ternary expression
func If(ok bool, trueValue interface{}, falseValue interface{}) interface{} {
	if ok {
		return trueValue
	}
	return falseValue
}

func GetContextByGin(c *gin.Context) context.Context {
	return c.Request.Context()
}

// RemoveEmptyStrings 删除 s 中空的 string
func RemoveEmptyStrings(s []string) (result []string) {
	for _, v := range s {
		if v == "" {
			continue
		}
		result = append(result, v)
	}
	return
}

func Title(s string) string {
	return cases.Title(language.English).String(s)
}

func Try(f func() error, tryCount ...int) (err error) {
	var maxTry = 3
	if len(tryCount) != 0 && tryCount[0] > 0 {
		maxTry = tryCount[0]
	}
	var try int
	for try < maxTry {
		try++
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}

func TryReturn(f func() (result interface{}, err error), tryCount ...int) (result interface{}, err error) {
	var maxTry = 3
	if len(tryCount) != 0 && tryCount[0] > 0 {
		maxTry = tryCount[0]
	}
	var try int
	for try < maxTry {
		try++
		result, err = f()
		if err == nil {
			return result, nil
		}
	}
	return nil, err
}

func ScheduledTask(ctx context.Context, f func(), interval time.Duration) {
	defer RecoverRunForever("ScheduledTask error", func() { ScheduledTask(ctx, f, interval) }, time.Second*20, true)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			f()
		case <-ctx.Done():
			return
		}
	}
}
