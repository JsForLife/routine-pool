package routinepool

import (
	"sync"
	"testing"
)

func TestPoolRoutine(t *testing.T) {
	// 创建一个协程池
	pool := NewPool(3, 1000)
	// 启动协程池
	if err := pool.Start(); err != nil {
		t.Error(err)
		return
	}

	// A 组任务
	tasksA := []func() (interface{}, error){
		func() (interface{}, error) { return "Task A1 result", nil },
		func() (interface{}, error) { return "Task A2 result", nil },
		func() (interface{}, error) { return "Task A3 result", nil },
	}

	// B 组任务
	tasksB := []func() (interface{}, error){
		func() (interface{}, error) { return "Task B1 result", nil },
		func() (interface{}, error) { return "Task B2 result", nil },
		func() (interface{}, error) { return "Task B3 result", nil },
	}

	tests := [][]func() (interface{}, error){
		tasksA,
		tasksB,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// 开启两个协程，各自向协程池添加任务和接收结果
	for _, tasks := range tests {
		go func(tasks []func() (interface{}, error)) {
			defer wg.Done()

			// 创建 future 对象，用于接收结果
			future := pool.newFuture(len(tasks))
			defer future.Close()

			// 向协程池添加任务
			for i, task := range tasks {
				if err := pool.AddTask(future, i, task); err != nil {
					t.Error(err)
					return
				}
			}

			// 接收结果
			results := make(map[int]Result, len(tasksA))
			for i := 0; i < len(tasks); i++ {
				// 通过 future 中的结果通道来保证这些结果属于当前调用方
				result := <-future.Result
				// 按任务的添加顺序采集结果，使出入参顺序一致
				results[result.index] = result
			}
			for i := 0; i < len(tasks); i++ {
				t.Log(results[i])
			}
		}(tasks)
	}
	wg.Wait()

	// 关闭协程池。也可以不关闭，常驻内存，下次直接用。
	if err := pool.Stop(); err != nil {
		t.Error(err)
	}
}
