package timewheel

import (
	"container/list"
	"log"
	"reflect"
	"runtime"
	"time"
)

// Job Func callback delay
type Job func(interface{})

// Task task to feed Job
type Task struct {
	delay  time.Duration
	circle int
	key    interface{} // task unique key to find the task-added
	data   interface{} // param
	cyclic bool        // is cyclic task
}

// TimeWheel golang task timewheel
type TimeWheel struct {
	interval          time.Duration // ticket move time interval
	ticker            *time.Ticker
	slots             []*list.List
	timer             map[interface{}]int
	currentPos        int
	slotNum           int
	job               Job
	addTaskChannel    chan Task
	removeTaskChannel chan interface{}
	stopChannel       chan bool
}

// New create a new timewheel
// @interval - ticket time interval
// @slotNum - num of timewheel's slot per cycle
// @job - job to run
func New(interval time.Duration, slotNum int, job Job) *TimeWheel {
	if interval <= 0 || slotNum <= 0 || job == nil {
		return nil
	}
	tw := &TimeWheel{
		interval:          interval,
		ticker:            time.NewTicker(interval),
		slots:             make([]*list.List, slotNum),
		timer:             make(map[interface{}]int),
		currentPos:        0,
		slotNum:           slotNum,
		job:               job,
		addTaskChannel:    make(chan Task),
		removeTaskChannel: make(chan interface{}),
		stopChannel:       make(chan bool),
	}
	tw.initSlots()
	return tw
}

// Start start the tw
func (tw *TimeWheel) Start() {
	go func() {
		for {
			select {
			case <-tw.ticker.C:
				tw.ticketHandle()
			case task := <-tw.addTaskChannel:
				tw.addTask(&task)
			case key := <-tw.removeTaskChannel:
				tw.removeTask(key)
			case <-tw.stopChannel:
				tw.ticker.Stop()
				return
			}
		}
	}()
	log.Printf(`###  timewheel started: slotsNum[%v],slotInterval[%v],job[%v]  ####`, tw.slotNum, tw.interval, runtime.FuncForPC(reflect.ValueOf(tw.job).Pointer()).Name())
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}
func (tw *TimeWheel) ticketHandle() {
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
}
func (tw *TimeWheel) addTask(task *Task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	tw.slots[pos].PushBack(task)

	if task.key != nil {
		tw.timer[task.key] = pos
	}
}

// AddTask add task to be done
// @ delay -  duration for delay task , can be bigger than slotNum
// @ cyclic - weather the task to be done forever
func (tw *TimeWheel) AddTask(delay time.Duration, cyclic bool, key interface{}, data interface{}) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- Task{delay: delay, key: key, data: data, cyclic: cyclic}
}
func (tw *TimeWheel) removeTask(key interface{}) {
	position, ok := tw.timer[key]
	if !ok {
		return
	}
	// get link of the special slots
	l := tw.slots[position]
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timer, task.key)
			l.Remove(e)
		}

		e = e.Next()
	}

}

// RemoveTask remove task
func (tw *TimeWheel) RemoveTask(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}
func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		go tw.job(task.data)
		next := e.Next()
		l.Remove(e)
		if task.key != nil {
			delete(tw.timer, task.key)
		}
		// if task is cyclic , push  the task to the  link
		if task.cyclic {
			position := tw.getPreviousTickIndex() + 1
			tw.slots[position].PushFront(task)
			if task.key != nil {
				tw.timer[task.key] = position
			}
		}
		e = next
	}
}
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	circle = int(delaySeconds / intervalSeconds / tw.slotNum)
	pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum
	return
}

func (tw *TimeWheel) getPreviousTickIndex() int {
	cti := tw.currentPos
	if 0 == cti {
		return tw.slotNum - 1
	}
	return cti - 1
}
