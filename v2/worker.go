package workers

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// WorkerPool struct
type WorkerPool struct {
	taskQueue   chan Task
	workerQueue chan bool

	opts Opts
}

// Task struct
type Task struct {
	Name string
	Do   func(w *WorkerPool)
}

// Opts options for logging, etc
type Opts struct {
	Name      string
	Silent    bool
	Verbose   bool
	QueueSize int
}

// New will create new workers package
// maxWorker will define the maximum worker available
func New(maxWorker int, opts ...Opts) *WorkerPool {
	var opt Opts
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = Opts{
			Name:      uuid.New().String(),
			Silent:    false,
			Verbose:   false,
			QueueSize: 100,
		}
	}
	wp := &WorkerPool{
		taskQueue: make(chan Task, opt.QueueSize),
		opts:      opt,
	}
	wp.startDispatcher(maxWorker)
	return wp
}

// StoreTask will adding a task into queue
func (wp *WorkerPool) StoreTask(task Task) {
	wp.taskQueue <- task
}

func (wp *WorkerPool) startDispatcher(nworkers int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	wp.workerQueue = make(chan bool, nworkers)

	if !wp.opts.Silent {
		log.Println("Starting", nworkers, "workers:", wp.opts.Name)
	}
	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		if wp.opts.Verbose {
			log.Println("Starting worker:", wp.opts.Name, i+1)
		}
		wp.workerQueue <- true
	}

	if !wp.opts.Silent {
		log.Println("Finished starting", nworkers, "workers:", wp.opts.Name)
	}

	go func(wp *WorkerPool) {
		for task := range wp.taskQueue {
			if wp.opts.Verbose {
				log.Println("Received work request:", wp.opts.Name)
			}
			go func(task Task) {
				<-wp.workerQueue
				if wp.opts.Verbose {
					log.Println("Dispatching work request:", wp.opts.Name)
				}
				task.Do(wp)
				wp.workerQueue <- true
			}(task)
		}
	}(wp)
}

// Pause task for duration and give control to other task
// To be called from inside a task
func (wp *WorkerPool) Pause(d time.Duration) {
	wp.workerQueue <- true
	time.Sleep(d)
	<-wp.workerQueue
}

// Yield to other task
// To be called from inside a task
func (wp *WorkerPool) Yield() {
	wp.workerQueue <- true
	<-wp.workerQueue
}
