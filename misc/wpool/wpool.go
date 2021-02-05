package wpool

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	bigTask uint32 = 500
)

// Pool maintains multiple workers (goroutine), which can
// be executed concurrently to perform tasks.
// The number of workers is controlled. When all workers
// are busy, new workers will be created, but the total
// number of workers will not exceed the preset threshold.
// All workers will always exist and can perform tasks
// repeatedly, and they will not be destroyed until the
// entire pool is closed.
//
// When there are many tasks that need to be executed
// concurrently, it is a good practice to use Pool.
// It can reuse workers to perform tasks and reduce the
// overhead of creating and destroying workers. And using
// Pool can avoid creating too many or too few workers,
// it let the pool automatically expand the number of
// workers to improve the efficiency of task execution.
//
// Pool can receive an unlimited number of tasks (like a
// daemon) or a limited number of tasks (like WaitGroup),
// see the "Total" and "Wait" documentation for details.
type Pool struct {
	// send tasks to the workers.
	tch chan *task

	// Used to control the opening/closing of the pool.
	// When ctx is cancelled, the entire pool will be closed.
	ctx      context.Context
	doneFunc context.CancelFunc

	// How many tasks are currently being executed.
	running int32

	// The number of concurrently executing workers and
	// the maximum number of workers. Note that worker
	// should always be less than or equal to maxWorker.
	worker    int32
	maxWorker int32

	// total represents the total number of tasks to be
	// executed. When the number of tasks executed is
	// equal to or exceeds total, the pool will stop.
	// A total of 0 means that the pool can perform an
	// unlimited number of tasks. cnt represents the
	// number of tasks currently and successfully executed.
	total uint32
	cnt   uint32

	// Once an error occurs during task execution, stop
	// the entire pool and assign the error to this field.
	err error

	// If it is 1, it means the pool has been closed.
	done uint32

	// Task function performed by the pool by default.
	// It is passed in when it is created, and action
	// can also be passed in when the task is submitted.
	action reflect.Value

	// use to init tch
	initOnce sync.Once
}

// New creates a default empty daemon pool, which can
// receive and execute unlimited tasks. After calling
// "Wait", it will block forever until a task returns an
// error. If you want the pool to close after performing
// a limited number of tasks, you need to call "Total".
// If you want to manually close the pool externally
// (for example, set the timeout for the pool), you need
// to call "NewWithContext" to create the pool.
//
// By default, the maximum number of workers is the
// current number of CPU cores. If you want to change
// this parameter, you need to call "Worker".
func New() *Pool {
	return NewWithContext(context.Background())
}

// NewWithContext creates a default daemon pool just like
// "New". The difference is that it uses the context.Context
// passed in from outside to initialize the pool. After the
// external Context is closed, the entire pool will be closed.
// Through this Context, the ability to forcibly close the
// pool externally is given, for example, when the pool execution
// times out, it is forced to close.
func NewWithContext(ctx context.Context) *Pool {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &Pool{
		maxWorker: int32(runtime.NumCPU()),

		ctx:      cancelCtx,
		doneFunc: cancelFunc,
	}
}

// Total sets the maximum number of tasks performed by
// the pool. If the number of tasks performed by the pool
// exceeds this value, the entire pool will be closed
// directly.
func (p *Pool) Total(n int) *Pool {
	if n <= 0 {
		panic("total must > 0")
	}
	p.total = uint32(n)
	return p
}

// Worker sets the maximum number of workers that pool
// will created. Note that this is not a fixed number of
// workers, the pool will dynamically create workers
// based on the submitted tasks, but the total number
// of creations will not exceed this number.
//
// By default, the maximum number of workers is equal to
// the number of CPU cores of the current machine.
func (p *Pool) Worker(n int) *Pool {
	if n <= 0 {
		panic("worker must > 0")
	}
	p.maxWorker = int32(n)
	return p
}

// Action sets the default action function of the pool.
// The action must be a function type, and its return
// value is either empty or an "error". If the return
// value of the function is empty, it means that the
// task will never return an error.
//
// Calling "SubmitArgs" when submitting a task will execute
// the action set here. Note that this function must be
// called before calling SubmitArgs, otherwise undefined
// behavior will occur.
//
// For example, the following two function types are
// valid:
//    func(a, b int)
//    func() error
//    func(s string, idx int) error
// But the following functions are invalid and will
// cause panic:
//    func(a, b int) (int, error)
//    func() (bool, error)
//
// There are no restrictions on the action parameters,
// but when submitting the task later, the parameters
// passed in must correspond to the action here.
func (p *Pool) Action(action interface{}) *Pool {
	p.action = reflect.ValueOf(action)
	return p
}

// Do uses the action function to create tasks and
// submit them to the pool for execution. If the
// action function returns an error, it will cause
// the entire pool to close.
//
// Do will return immediately. The order in which Do
// is called and the order in which actions are
// actually executed may not be consistent.
// Do not call Do on a closed pool, and Do not call
// Do after Wait, otherwise "blocking forever" or
// panic may occur.
//
// Do itself is non-concurrently safe, it should
// be called sequentially.
func (p *Pool) Do(action func() error) {
	p.submit(reflect.ValueOf(action), nil)
}

// SubmitArgs uses the default action to submit the
// task. If the default action function has parameters,
// the "args" must correspond to the function parameters.
// You must call "Action" before submitting a task
// in this way.
//
// SubmitArgs will return immediately. The order
// in which SubmitArgs is called and the order
// in which actions are  actually executed may
// not be consistent. Do not call SubmitArgs on
// a closed pool, and do not call SubmitArgs after
// Wait, otherwise "blocking forever" or panic
// may occur.
//
// SubmitArgs itself is non-concurrently safe, it should
// be called sequentially.
func (p *Pool) SubmitArgs(args ...interface{}) {
	p.submit(p.action, args)
}

// Submit uses the given function to submit the task,
// and calling Submit does not need to call "Action"
// first. Use Submit to allow different tasks to perform
// different functions. The remaining behavior of Submit
// is the same as SubmitArgs.
//
// Submit will return immediately. The order
// in which Submit is called and the order
// in which actions are  actually executed may
// not be consistent. Do not call Submit on
// a closed pool, and do not call Submit after
// Wait, otherwise "blocking forever" or panic
// may occur.
//
// Submit itself is non-concurrently safe, it should
// be called sequentially.
func (p *Pool) Submit(action interface{}, args ...interface{}) {
	p.submit(reflect.ValueOf(action), args)
}

// Wait blocks until the pool is closed, and releases
// the pool's resources after the pool is closed.
// Returns error generated during the execution of the pool.
// After Wait returns, the pool has been released and can no
// longer call any of its functions (including Wait itself).
// Therefore, it can be said that Wait should be the last
// function called by the pool in any case (just like the
// Close function of some resource objects).
// If the following conditions occur, the pool will be closed
// and Wait will stop blocking:
//  1. An error occurred in the execution of a task.
//	2. The total parameter is set, and the number of executed
//     tasks reaches total (see "Total").
//	3. The pool is created by "NewWithContext", and the passed
//     context is done.
//
func (p *Pool) Wait() error {
	<-p.ctx.Done()
	close(p.tch)
	return p.err
}

func (p *Pool) submit(action reflect.Value, args []interface{}) {
	p.initOnce.Do(p.init)
	if atomic.LoadUint32(&p.done) == 1 {
		return
	}

	if p.worker == 0 || p.isBusy() {
		p.worker += 1
		go p.work()
	}

	p.tch <- &task{action: action, args: args}
}

func (p *Pool) init() {
	var n int
	switch {
	case p.total == 0:
		n = int(p.maxWorker)

	case p.total < bigTask:
		n = int(p.total)

	default:
		n = int(bigTask)
	}

	p.tch = make(chan *task, n)
}

func (p *Pool) isBusy() bool {
	return atomic.LoadInt32(&p.running) >= p.worker && p.worker < p.maxWorker
}

func (p *Pool) work() {
	isWorkerDone := false
	workDone := func() {
		isWorkerDone = true
		p.doneFunc()
	}
	for t := range p.tch {
		if isWorkerDone {
			continue
		}

		if atomic.LoadUint32(&p.done) == 1 {
			workDone()
			continue
		}

		atomic.AddInt32(&p.running, 1)
		err := t.exec()
		atomic.AddInt32(&p.running, -1)
		if err != nil {
			p.err = err
			atomic.StoreUint32(&p.done, 1)
			workDone()
			continue
		}
		cnt := atomic.AddUint32(&p.cnt, 1)
		if p.total != 0 && cnt >= p.total {
			atomic.StoreUint32(&p.done, 1)
			workDone()
		}
	}
}

type task struct {
	action reflect.Value
	args   []interface{}
}

func (t *task) exec() error {
	in := make([]reflect.Value, len(t.args))
	for idx, arg := range t.args {
		in[idx] = reflect.ValueOf(arg)
	}
	rets := t.action.Call(in)
	switch len(rets) {
	case 0:
		return nil

	case 1:
		ret := rets[0]
		if ret.IsNil() {
			return nil
		}

		return ret.Interface().(error)

	default:
		panic("invalid action returns")
	}
}
