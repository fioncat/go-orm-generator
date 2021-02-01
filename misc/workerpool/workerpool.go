package workerpool

import (
	"sync"
)

// WorkFunc represents the specific function type that
// the worker runs.
type WorkFunc func(task interface{}) error

// Pool represents a pool that stores a fixed number of
// workers. It can have multiple implementations.
// Tasks can be initiated to the pool, and the tasks will
// be assigned to a worker to run.
type Pool interface {
	// Start opens the pool. Before calling Start, there
	// is no worker in the pool and no task can be received.
	// So must call this method before calling Do.
	Start()

	// Do sends a task to the pool. The pool should schedule
	// it to be executed by a worker. In the case of concurrency,
	// Do theoretically returns immediately.
	Do(task interface{})

	// Wait blocks until all sent tasks are executed, and
	// returns to errors generated during execution. Wait
	// should clean up the resources occupied by the pool.
	Wait() error
}

// pool is the implementation of the concurrent version of
// the Pool interface. Its multiple workers will asynchronously
// preempt a channel to obtain tasks. Do will send tasks
// directly to this channel for workers to preempt.
type pool struct {
	taskCh chan interface{}

	doneCh chan int
	errCh  chan error

	nWorker int
	nTask   int

	stopCh chan error

	workFunc WorkFunc

	isClosed bool
	closeMu  sync.RWMutex
}

// oneTask is an implementation of the Pool interface
// with only one task. In fact, it simply executes
// workFunc once.
type oneTask struct {
	workFunc WorkFunc
	task     interface{}
}

func (s *oneTask) Start()              {}
func (s *oneTask) Do(task interface{}) { s.task = task }
func (s *oneTask) Wait() error         { return s.workFunc(s.task) }

// oneWorker is a Pool interface implementation with
// only one worker but multiple tasks. It will actually
// perform tasks sequentially.
type oneWorker struct {
	workFunc WorkFunc
	err      error
}

func (s *oneWorker) Start() {}

func (s *oneWorker) Do(task interface{}) {
	// skip if any error exists.
	if s.err != nil {
		return
	}
	s.err = s.workFunc(task)
}

func (s *oneWorker) Wait() error {
	return s.err
}

// New creates a new Pool. The specific pool implementation
// structure is determined internally and externally does
// not need to be concerned.
func New(nTask, nWorker int, workFunc WorkFunc) Pool {
	if nTask == 1 {
		return &oneTask{workFunc: workFunc}
	}
	if nWorker == 1 {
		return &oneWorker{workFunc: workFunc}
	}
	return &pool{
		taskCh:   make(chan interface{}, nWorker),
		doneCh:   make(chan int, nWorker),
		errCh:    make(chan error, 5),
		nWorker:  nWorker,
		nTask:    nTask,
		stopCh:   make(chan error, 5),
		workFunc: workFunc,
	}
}

func (p *pool) Start() {
	for i := 0; i < p.nWorker; i++ {
		w := &worker{
			taskCh: p.taskCh,
			doneCh: p.doneCh,
			errCh:  p.errCh,
			work:   p.workFunc,
			pool:   p,
		}
		go w.start()
	}
	go p.listen()
}

func (p *pool) Do(task interface{}) {
	p.closeMu.RLock()
	defer p.closeMu.RUnlock()
	if p.isClosed {
		// If the pool is closed,
		return
	}
	p.taskCh <- task
}

func (p *pool) listen() {
	cnt := 0
	sentStop := false
	for {
		if p.isClosed {
			return
		}
		select {
		case done := <-p.doneCh:
			cnt += done
			if cnt >= p.nTask {
				if !sentStop {
					p.closeMu.RLock()
					if !p.isClosed {
						p.stopCh <- nil
						sentStop = true
					}
					p.closeMu.RUnlock()
				}
			}

		case err := <-p.errCh:
			if !sentStop {
				p.closeMu.RLock()
				if !p.isClosed {
					p.stopCh <- err
					sentStop = true
				}
				p.closeMu.RUnlock()
			}
		}
	}
}

func (p *pool) closeCh() {
	p.closeMu.Lock()
	defer p.closeMu.Unlock()
	if p.isClosed {
		return
	}
	close(p.errCh)
	close(p.stopCh)
	close(p.doneCh)
	close(p.taskCh)
	p.isClosed = true
}

func (p *pool) Wait() error {
	defer p.closeCh()
	for {
		err := <-p.stopCh
		return err
	}
}

// worker that execute workFunc concurrently, will
// get the tasks to be executed from taskCh.
type worker struct {
	work WorkFunc

	taskCh chan interface{}
	doneCh chan int
	errCh  chan error

	pool *pool
}

func (w *worker) start() {
	for {
		task, ok := <-w.taskCh
		if !ok || task == nil {
			return
		}
		// If the pool is already closed, stop the worker.
		if w.pool.isClosed {
			return
		}
		err := w.work(task)
		if err != nil {
			// Work with error, try to notify pool.
			w.pool.closeMu.RLock()
			if w.pool.isClosed {
				return
			}
			// notify error
			w.errCh <- err
			w.pool.closeMu.RUnlock()
		}
		w.pool.closeMu.RLock()
		if w.pool.isClosed {
			// Work done, but the pool is closed,
			// stop the worker directly.
			w.pool.closeMu.RUnlock()
			return
		}
		w.doneCh <- 1
		w.pool.closeMu.RUnlock()
	}
}
