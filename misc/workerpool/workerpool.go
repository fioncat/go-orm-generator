package workerpool

import (
	"sync"

	"github.com/fioncat/go-gendb/misc/log"
)

type WorkFunc func(task interface{}) error

type Pool interface {
	Start()
	Do(task interface{})
	Wait() error
}

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

type single struct {
	workFunc WorkFunc
	task     interface{}
}

func (s *single) Start()              {}
func (s *single) Do(task interface{}) { s.task = task }
func (s *single) Wait() error         { return s.workFunc(s.task) }

func New(nTask, nWorker int, workFunc WorkFunc) Pool {
	if nTask == 1 {
		return &single{workFunc: workFunc}
	}
	return &pool{
		taskCh:   make(chan interface{}, nWorker),
		doneCh:   make(chan int, nWorker),
		errCh:    make(chan error, 1),
		nWorker:  nWorker,
		nTask:    nTask,
		stopCh:   make(chan error, 1),
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
	for {
		if p.isClosed {
			return
		}
		select {
		case done := <-p.doneCh:
			cnt += done
			if cnt >= p.nTask {
				p.stopCh <- nil
				return
			}

		case err := <-p.errCh:
			p.stopCh <- err
			return
		}
	}
}

func (p *pool) Wait() error {
	defer func() {
		p.closeMu.Lock()
		close(p.errCh)
		close(p.stopCh)
		close(p.doneCh)
		close(p.taskCh)
		p.isClosed = true
		p.closeMu.Unlock()
	}()
	for {
		err := <-p.stopCh
		return err
	}
}

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
			// If the pool is already closed, log
			// this error and return.
			w.pool.closeMu.RLock()
			defer w.pool.closeMu.RUnlock()
			if w.pool.isClosed {
				log.Errorf("unhandle error from worker: %v", err)
				return
			}
			// notify error
			w.errCh <- err
			return
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
