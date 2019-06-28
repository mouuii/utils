package timewheel

import (
	"log"
	"time"
)

// DefaultTimeWheel - DEFAULTSLOTS seconds
var DefaultTimeWheel *TimeWheel

const (
	// DEFAULTSLOTS  default timewheel slots number
	DEFAULTSLOTS = 60
)

func init() {
	DefaultTimeWheel = New(time.Second, DEFAULTSLOTS, say)
}
func say(args interface{}) {
	log.Printf("default timewheel say %v\n", args)
}

// QuickStart start a timewheel fast
// example timewheel.QuickStart(3, func(p interface{}) { fmt.Println("ddd", p) }, "xxx", true)
func QuickStart(slotNum int, job Job, param interface{}, cyclic bool) {
	quickTime := New(time.Second, slotNum, job)
	quickTime.Start()
	quickTime.AddTask(time.Second, cyclic, nil, param)
	return
}
