package globals

import (
	"sync"
)

type Counter struct {
	lock           sync.Mutex
	noOfCalls      int32
	noOfCallObject int32
}

var counters *Counter

func InitCounter() {
	counters = new(Counter)
}

func GetNoOfCalls() int32 {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	return counters.noOfCalls
}

func IncrementNoOfCalls() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfCalls++
}

func DecrementNoOfCalls() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfCalls--
}

func GetNoOfCallObject() int32 {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	return counters.noOfCalls
}

func IncrementNoOfCallObject() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfCallObject++
}

func DecrementNoOfCallObject() {
	counters.lock.Lock()
	defer counters.lock.Unlock()
	counters.noOfCallObject--
}
