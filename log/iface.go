package log

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robolivable/beaves/config"
)

type memo struct {
	Exp   int64
	Count int64
}

var memoizeLogs = map[string]memo{}
var memoizeLock sync.Mutex

func println(msg string, args ...any) {
	if !config.RuntimeConfig.Log.Enabled {
		return
	}
	fmt.Printf(msg+"\n", args...)
}

func Info(msg string, args ...any) {
	println("info: "+msg, args...)
}

func InfoMemoize(msg string, args ...any) {
	memoizeLock.Lock()
	defer memoizeLock.Unlock()
	log := strings.ToLower(fmt.Sprintf(msg, args...))
	m := memoizeLogs[log]
	if m.Exp != 0 && m.Exp > time.Now().UnixMilli() {
		m.Count += 1
		memoizeLogs[log] = m
		return
	}
	count := m.Count
	m = memo{
		Exp:   time.Now().Add(time.Duration(60) * time.Second).UnixMilli(),
		Count: 0,
	}
	memoizeLogs[log] = m
	Info(fmt.Sprintf("[%d, %d]", time.Now().UnixMilli(), count)+" "+msg, args...)
}

func Error(msg string, args ...any) {
	println("error: "+msg, args...)
}
