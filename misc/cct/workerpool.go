package cct

import "github.com/fioncat/go-gendb/misc/log"

type WorkFunc func(task interface{}) error

type Pool struct {
	taskCh chan interface{}

	doneCh chan int
	errCh  chan error

	nWorker int
	nTask   int

	stopCh chan error

	workFunc WorkFunc
}

func NewPool(nTask, nWorker int, workFunc WorkFunc) *Pool {
	return &Pool{
		taskCh:   make(chan interface{}, nWorker),
		doneCh:   make(chan int, nWorker),
		errCh:    make(chan error, 1),
		nWorker:  nWorker,
		nTask:    nTask,
		stopCh:   make(chan error, 1),
		workFunc: workFunc,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.nWorker; i++ {
		w := &worker{
			taskCh: p.taskCh,
			doneCh: p.doneCh,
			errCh:  p.errCh,
			work:   p.workFunc,
		}
		go w.start()
	}
	go p.listen()
}

func (p *Pool) Do(task interface{}) {
	p.taskCh <- task
}

func (p *Pool) listen() {
	cnt := 0
	for {
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

func (p *Pool) Wait() error {
	defer close(p.errCh)
	defer close(p.doneCh)
	defer close(p.taskCh)
	defer close(p.stopCh)
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
}

func (w *worker) start() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("worker recover: %v", r)
			return
		}
	}()
	for {
		task, ok := <-w.taskCh
		if !ok || task == nil {
			return
		}
		err := w.work(task)
		if err != nil {
			w.errCh <- err
			return
		}
		w.doneCh <- 1
	}
}
