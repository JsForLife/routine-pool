package routinepool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Task 任务结构体，包含任务的唯一索引、具体任务函数和响应通道
type Task struct {
	index      int
	fn         func() (interface{}, error)
	responseCh chan Result
}

type Result struct {
	index int
	value interface{}
	err   error
}

type Future struct {
	Result chan Result
}

// Pool 协程池结构体
type Pool struct {
	numWorkers int
	taskCh     chan Task
	workers    []*worker
	closed     atomic.Bool
	wg         sync.WaitGroup
	shutdownCh chan struct{}
}

// NewPool 创建一个新的协程池
func NewPool(numWorkers int, taskChSize int) *Pool {
	if taskChSize <= 0 {
		taskChSize = 1
	}
	p := &Pool{
		numWorkers: numWorkers,
		taskCh:     make(chan Task, taskChSize),
		workers:    make([]*worker, numWorkers),
		shutdownCh: make(chan struct{}),
	}
	p.closed.Store(true)

	for i := 0; i < numWorkers; i++ {
		w := &worker{
			id:        i,
			owner:     p,
			requestCh: make(chan Task),
			shutdown:  make(chan struct{}),
		}
		p.workers[i] = w
	}
	return p
}

// Start 启动协程池
func (p *Pool) Start() error {
	if !p.closed.CompareAndSwap(true, false) {
		return fmt.Errorf("the pool is already running")
	}
	// wg 用于等待 dispatch 和 所有 worker 退出
	p.wg.Add(p.numWorkers + 1)
	for _, w := range p.workers {
		go w.start()
	}
	go p.dispatch()
	return nil
}

// 新建每个调用方独占的结果通道
func (p *Pool) newFuture(size int) *Future {
	return &Future{
		Result: make(chan Result, size),
	}
}

// AddTask 向协程池添加任务，接收调用方传入的响应通道
func (p *Pool) AddTask(future *Future, index int, fn func() (interface{}, error)) error {
	if p.closed.Load() {
		return fmt.Errorf("the pool is closed")
	}
	task := Task{index: index, fn: fn, responseCh: future.Result}
	p.taskCh <- task
	return nil
}

// Stop 关闭协程池
func (p *Pool) Stop() error {
	if !p.closed.CompareAndSwap(false, true) {
		return fmt.Errorf("the pool is already closed")
	}
	close(p.taskCh)
	close(p.shutdownCh)
	for _, w := range p.workers {
		close(w.shutdown)
	}
	p.wg.Wait()
	return nil
}

// dispatch 调度任务给工作协程
func (p *Pool) dispatch() {
	defer p.wg.Done()
	taskIndex := 0
	for {
		select {
		case task, ok := <-p.taskCh:
			if !ok {
				// 任务通道关闭，可能是 Stop 函数调用导致
				return
			}
			workerIndex := taskIndex % p.numWorkers
			p.workers[workerIndex].requestCh <- task
			taskIndex++
		case <-p.shutdownCh:
			return
		}
	}
}

// 关闭接收结果的通道
func (f *Future) Close() {
	close(f.Result)
}
