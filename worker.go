package workers

import (
	"log"

	"github.com/google/uuid"
)

// Workers struct
type Workers struct {
	workers     []worker
	workQueue   chan WorkerTask
	workerQueue chan chan WorkerTask

	opts Opts
}

// WorkerTask struct
type WorkerTask struct {
	Name string
	Do   func()
}

type Opts struct {
	Silent  bool
	Verbose bool
}

// New will create new workers package
// maxWorker will define the maximum worker available
func New(maxWorker int, opts ...Opts) *Workers {
	var opt Opts
	if len(opts) > 0 {
		opt = opts[0]
	}
	w := &Workers{
		workQueue: make(chan WorkerTask, 100),
		opts:      opt,
	}
	w.startDispatcher(maxWorker)
	return w
}

// StoreTask will adding a task into queue
func (w *Workers) StoreTask(task WorkerTask) {
	w.workQueue <- task
}

// Stop stop all running task
func (w *Workers) Stop() {
	for _, w := range w.workers {
		w.stop()
	}
}

func (w *Workers) startDispatcher(nworkers int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	w.workerQueue = make(chan chan WorkerTask, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		if !w.opts.Silent {
			log.Println("Starting workers", i+1)
		}
		work := worker{
			ID:          uuid.New().String(),
			Work:        make(chan WorkerTask),
			WorkerQueue: w.workerQueue,
			QuitChan:    make(chan bool)}
		w.workers = append(w.workers, work)
		work.start()
	}

	go func(wk *Workers) {
		for {
			select {
			case workQ := <-wk.workQueue:
				if w.opts.Verbose {
					log.Println("Received work requeust")
				}
				go func() {
					work := <-wk.workerQueue
					if w.opts.Verbose {
						log.Println("Dispatching work request")
					}
					work <- workQ
				}()
			}
		}
	}(w)
}

// worker struct
type worker struct {
	ID          string
	Work        chan WorkerTask
	WorkerQueue chan chan WorkerTask
	QuitChan    chan bool
}

// This function "starts" the workers by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *worker) start() {
	go func(wk *worker) {
		for {
			// Add ourselves into the workers queue.
			wk.WorkerQueue <- wk.Work

			select {
			case work := <-wk.Work:
				// Receive and execute the function
				work.Do()
			case <-wk.QuitChan:
				return
			}
		}
	}(w)
}

// Stop tells the workers to stop listening for work requests.
// Note that the workers will only stop *after* it has finished its work.
func (w *worker) stop() {
	w.QuitChan <- true
}
