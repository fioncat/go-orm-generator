package gpool

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
)

type Pool interface {
	Start()
	Do(args ...interface{})
	Wait() error
}

func New(ctx context.Context, worker, total int, action interface{}) Pool {
	if worker == 1 || total == 1 {
		return &sequence{action: newAction(action), ctx: ctx}
	}
	return newPool(ctx, worker, total, action)
}

func newAction(actionFunc interface{}) reflect.Value {
	v := reflect.ValueOf(actionFunc)
	if v.Kind() != reflect.Func {
		panic("action is not a function")
	}
	return v
}

func callAction(action reflect.Value, args []interface{}) error {
	in := make([]reflect.Value, len(args))
	for idx, arg := range args {
		in[idx] = reflect.ValueOf(arg)
	}

	rets := action.Call(in)
	switch len(rets) {
	case 0:
		// Never returns error
		return nil

	case 1:
		// returns error
		ret := rets[0]
		if ret.IsNil() {
			return nil
		}
		return ret.Interface().(error)

	default:
		// invalid returns
		panic("invalid returns for the action")
	}
}

type sequence struct {
	status uint32
	err    error
	action reflect.Value
	ctx    context.Context

	done chan struct{}
}

func (sq *sequence) Start() {
	atomic.StoreUint32(&sq.status, statusRun)
	sq.done = make(chan struct{}, 1)
	go sq.listen()
}

func (sq *sequence) listen() {
	for {
		select {
		case <-sq.ctx.Done():
			atomic.StoreUint32(&sq.status, statusDeadline)
			sq.err = ContextDeadlineExceeded

		case <-sq.done:
			close(sq.done)
			return
		}
	}
}

func (sq *sequence) Do(args ...interface{}) {
	if atomic.LoadUint32(&sq.status) == statusRun {
		err := callAction(sq.action, args)
		if err != nil {
			atomic.StoreUint32(&sq.status, statusFail)
			sq.err = err
		}
	}
}

func (sq *sequence) Wait() error {
	sq.done <- struct{}{}
	return sq.err
}

const (
	statusRun      uint32 = 0
	statusDone     uint32 = 1
	statusFail     uint32 = 2
	statusDeadline uint32 = 3
	statusDead     uint32 = 4

	bigTask uint32 = 500
)

var ContextDeadlineExceeded = errors.New("context deadline exceeded")

type pool struct {
	nw uint32
	nt uint32

	action reflect.Value

	invoke   chan []interface{}
	feedback chan error
	stop     chan struct{}

	ctx context.Context

	closeMu sync.RWMutex

	status uint32
	err    error
}

func newPool(ctx context.Context, worker, total int, action interface{}) *pool {
	p := new(pool)
	if worker <= 0 {
		panic("invalid worker")
	}
	if total < 0 {
		panic("invalid total")
	}
	p.nw = uint32(worker)
	p.nt = uint32(total)
	p.ctx = ctx
	p.action = newAction(action)

	return p
}

func (p *pool) Start() {
	var buffLen int
	if p.nt >= bigTask {
		buffLen = int(bigTask)
	} else {
		buffLen = int(p.nt)
	}

	p.invoke = make(chan []interface{}, buffLen)
	p.feedback = make(chan error, buffLen)

	p.stop = make(chan struct{}, 1)

	atomic.StoreUint32(&p.status, statusRun)

	var idx uint32
	for idx = 0; idx < p.nw; idx++ {
		go p.work()
	}

	go p.listen()
}

func (p *pool) work() {
	for {
		args, ok := <-p.invoke
		if !ok {
			return
		}
		err := callAction(p.action, args)
		p.closeMu.RLock()
		if atomic.LoadUint32(&p.status) != statusRun {
			// returns directly to prevent panic
			// because `feedback` has been closed.
			p.closeMu.RUnlock()
			return
		}
		p.feedback <- err
		p.closeMu.RUnlock()
	}
}

func (p *pool) Do(args ...interface{}) {
	p.closeMu.RLock()
	defer p.closeMu.RUnlock()
	if atomic.LoadUint32(&p.status) == statusRun {
		p.invoke <- args
	}
}

func (p *pool) listen() {
	var cnt uint32
mainLoop:
	for {
		select {
		case err := <-p.feedback:
			if err != nil {
				p.err = err
				atomic.StoreUint32(&p.status, statusFail)
				break mainLoop
			}
			cnt += 1
			if p.nt > 0 && cnt >= p.nt {
				atomic.StoreUint32(&p.status, statusDone)
				break mainLoop
			}

		case <-p.ctx.Done():
			atomic.StoreUint32(&p.status, statusDeadline)
			break mainLoop
		}
	}
	p.stop <- struct{}{}
}

func (p *pool) Close() {
	p.closeMu.Lock()
	p.status = statusDead
	close(p.invoke)
	close(p.feedback)
	close(p.stop)
	p.closeMu.Unlock()
}

func (p *pool) Wait() error {
	defer p.Close()
	<-p.stop
	switch p.status {
	case statusDone:
		return nil

	case statusFail:
		return p.err

	case statusDeadline:
		return ContextDeadlineExceeded
	}

	return nil
}
