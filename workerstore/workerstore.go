//Simple in-memory sync safe store for Task workers.
package workerstore

import (
	"sync"
	"time"
)

type Status string

const (
	Scheduled Status = "scheduled"
	Working          = "working"
	Done             = "done"
)

//Data we get from user
type Task struct {
	N   uint    `json:"n" validate:"min=0"`   //Number of elements in sequence
	D   float64 `json:"d"`                    //Delta = N(i + 1) - N(i)
	N1  float64 `json:"n1"`                   //First element
	I   float64 `json:"I" validate:"min=0"`   //Interval (seconds)
	TTL float64 `json:"TTL" validate:"min=0"` //Result store time (seconds)
}

//Additional fields
type Worker struct {
	Task
	NumInQueue   uint   `json:"numInQueue,omitempty"`
	CurrIter     uint   `json:"currIteration,omitempty"`
	ScheduleTime string `json:"scheduledTime"`
	StartTime    string `json:"startTime,omitempty"`
	EndTime      string `json:"endTime,omitempty"`
	Status       Status `json:"status,omitempty"`
}

type WorkerStore struct {
	mx sync.RWMutex

	scheduled   []Worker        //store for scheduled tasks
	workingdone map[int]*Worker //store for task being in progress or done
	max_working chan bool       //max number parrallel workers alowed to run
	done_chan   chan int        //channel for done workers id's
	next_id     int             //next key for workingdone
}

func NewWorkerStore(max_working uint) *WorkerStore {
	store := &WorkerStore{}
	store.workingdone = make(map[int]*Worker)
	store.max_working = make(chan bool, max_working)
	store.done_chan = make(chan int)
	store.next_id = 0
	return store
}

func (w *WorkerStore) AddTask(task Task) Worker {
	w.mx.Lock()
	defer w.mx.Unlock()
	worker := Worker{
		Task:         task,
		NumInQueue:   uint(len(w.scheduled) + 1),
		ScheduleTime: time.Now().UTC().Format(time.RFC3339Nano),
		Status:       Scheduled,
	}
	w.scheduled = append(w.scheduled, worker)
	return worker
}

//Gets done worker id from done_chan, waits TTL and removes worker
func (w *WorkerStore) waitTtl() {
	for {
		select {
		case id := <-w.done_chan:
			go func(id int) {
				defer w.mx.Unlock()
				defer delete(w.workingdone, id)
				time.Sleep(time.Millisecond * time.Duration(w.workingdone[id].TTL*1000))
				w.mx.Lock()
			}(id)
		}
	}
}

//Executes worker and sends it's id to done_chan then frees one place in max_working channel.
func (w *WorkerStore) executeWorker(id int) {
	worker := w.workingdone[id]
	worker.StartTime = time.Now().UTC().Format(time.RFC3339Nano)
	worker.Status = Working
	for i := 0; i < int(worker.N); i++ {
		worker.CurrIter += 1
		time.Sleep(time.Millisecond * time.Duration(worker.I*1000))
		worker.N1 += worker.D
	}
	worker.EndTime = time.Now().UTC().Format(time.RFC3339Nano)
	worker.Status = Done
	worker.CurrIter = 0
	w.done_chan <- id
	<-w.max_working
}

//Waits until there is free place in max_working channel then takes one place in channel
//Recalculates NumInQueue, sents first in queue worker to execution and removes it from queue
func (w *WorkerStore) StartWorkers() {
	defer close(w.done_chan)
	defer close(w.max_working)
	go w.waitTtl()
	for {
		if len(w.scheduled) > 0 {
			w.max_working <- true
			w.mx.Lock()
			for i := range w.scheduled {
				w.scheduled[i].NumInQueue = uint(i)
			}
			w.workingdone[w.next_id] = &w.scheduled[0]
			go w.executeWorker(w.next_id)
			w.next_id++
			w.scheduled = w.scheduled[1:]
			w.mx.Unlock()
		}
	}
}

func (w *WorkerStore) GetSortedTasks() []Worker {
	var res []Worker
	w.mx.RLock()
	defer w.mx.RUnlock()
	res = append(res, w.scheduled...)
	for _, val := range w.workingdone {
		if val.Status == Working {
			res = append(res, *val)
		}
	}
	for _, val := range w.workingdone {
		if val.Status == Done {
			res = append(res, *val)
		}
	}
	return res
}
