package util

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/spf13/cast"
)

// Panic if err != nil panic
func Panic(err error, msgs ...string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", strings.Join(msgs, ","), err))
		// var msg = "panic"
		// if len(msgs) != 0 && msgs[0] != "" {
		// 	msg = msgs[0]
		// }
		// log.Panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func Recover(msg string, showStack ...bool) {
	if err := recover(); err != nil {
		log.DPanic(fmt.Sprintf("%s: %s", msg, cast.ToString(err)))
		if len(showStack) != 0 && showStack[0] {
			log.Error(string(debug.Stack()))
		}
	}
}

func RecoverRunForever(msg string, f func(), interval time.Duration, showStack ...bool) {
	if err := recover(); err != nil {
		if errors.Is(err.(error), context.Canceled) ||
			strings.Contains(err.(error).Error(), "canceled") {
			log.Info("exit '%s'", msg)
			return
		}
		log.DPanic(fmt.Sprintf("%s: %s", msg, cast.ToString(err)))
		if len(showStack) != 0 && showStack[0] {
			log.Error(string(debug.Stack()))
		}
		time.Sleep(interval)
		f()
	}
	f()
}
