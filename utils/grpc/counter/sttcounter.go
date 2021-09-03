package counter


import (
	"sync"
)

type Counter struct {
	lock                    sync.Mutex
	noOfSTTConnections      int32
}

var counters Counter

func Get() int32 {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	return counters.noOfSTTConnections
}

func Increment() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfSTTConnections++
}

func Decrement() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfSTTConnections--
}