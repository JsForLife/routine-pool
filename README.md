# Go Routine Pool
This is a goroutine pool implemented in the Go language, aiming to provide a convenient way to manage and schedule multiple concurrent tasks, avoiding performance and resource management issues caused by unrestricted creation of `goroutines`.

```go
import "github.com/JsForLife/routine-pool"
```

This package requires Go 1.18.

## Features
- **Configurable Pool Size**: You can specify the number of worker goroutines when creating the goroutine pool.
- **Task Scheduling and Distribution**: Tasks will be automatically distributed to available worker goroutines, avoiding blocking between tasks.
- **Task Result Handling**: Using `Future` objects and result channels, you can easily receive and process the execution results of tasks.
- **Safe Shutdown Mechanism**: Provides a safe goroutine pool shutdown method to ensure that the goroutine pool is closed correctly after all tasks are completed.

## Usage Example
Here is a simple usage example showing how to create a goroutine pool, add tasks, and handle task results:
```go
package main

import (
    "fmt"
    "sync"

    "github.com/your_github_username/routinepool"
)

func main() {
    // Create a goroutine pool, specifying the number of worker goroutines and the task channel size (Make sure it's more than the number of tasks)
    pool := routinepool.NewPool(3, 10)

    // Start the goroutine pool
    if err := pool.Start(); err!= nil {
        fmt.Println(err)
        return
    }
    defer func() {
        // Close the goroutine pool if you want
        if err := pool.Stop(); err!= nil {
            fmt.Println(err)
        }
    }()

    // Define tasks
    tasks := []func() (interface{}, error){
        func() (interface{}, error) {
            return "Task 1 result", nil
        },
        func() (interface{}, error) {
            return "Task 2 result", nil
        },
        func() (interface{}, error) {
            return "Task 3 result", nil
        },
    }

    // Create a new Future object for receiving results
    future := pool.newFuture(len(tasks))
    defer future.Close()

    // Add tasks to the goroutine pool, i is the which use to match result
    for i, task := range tasks {
        if err := pool.AddTask(future, i, task); err!= nil {
            fmt.Println(err)
            return
        }
    }

    // Receive result
    for result := range future.Result {
        if result.err!= nil {
            fmt.Printf("Task %d failed: %v\n", result.index, result.err)
        } else {
            fmt.Printf("Task %d result: %v\n", result.index, result.value)
        }
    }
}
```

## Structure and Methods of the Goroutine Pool
### `Pool` Struct
The `Pool` struct represents a goroutine pool and contains the following important fields:
- `numWorkers`: The number of worker goroutines in the pool.
- `taskCh`: The task channel used to receive tasks to be executed.
- `workers`: A slice storing worker goroutines.
- `closed`: An `atomic.Bool` used to represent the closed state of the goroutine pool, ensuring atomic operations.
- `wg`: A `sync.WaitGroup` used to wait for all worker goroutines in the pool to complete.
- `shutdownCh`: A shutdown channel used to notify the goroutine pool to shut down

### Important Methods
- NewPool:
```go
func NewPool(numWorkers int, taskChSize int) *Pool
```
Creates a new goroutine pool, taking the number of worker goroutines and the task channel size as parameters. `taskChSize` should more than the number of tasks.

- Start:
```go
func (p *Pool) Start() error
```
Starts the goroutine pool. Returns an error if the pool is already running.
- newFuture:
```go
func (p *Pool) newFuture(size int) *Future
```
Creates an exclusive result channel for the caller, taking the buffer size of the result channel as a parameter.
- AddTask:
```go
func (p *Pool) AddTask(future *Future, index int, fn func() (interface{}, error)) error
```
Adds a task to the goroutine pool, taking a `Future` object, task index, and task function as parameters. Returns an error if the pool is closed.
- Stop:
```go
func (p *Pool) Stop() error
```
Closes the goroutine pool, waits for all worker goroutines to complete, and closes relevant channels.

### `worker` Struct
The `worker` struct represents a worker goroutine and contains the following important fields:
- `id`: The unique identifier of the worker goroutine.
- `owner`: A pointer to Pool, indicating the owning goroutine pool.
- `requestCh`: The channel for receiving tasks.
- `shutdown`: The shutdown channel used to notify the worker goroutine to shut down.

### Important Methods
- start:
```go
func (w *worker) start()
```
Starts the worker goroutine, receives tasks from `requestCh`, executes the tasks, and sends the results to the corresponding result channels.

### `Task` Struct
The `Task` struct represents a task and contains the following important fields:
- `index`: The unique index of the task.
- `fn`: The execution function of the task, which returns the result and a possible error.
- `responseCh`: The result channel used to store the task execution result.

### `Future` Struct
The `Future` struct represents a future result and contains the following important fields:
- `Result`: The result channel used to receive the execution result of the task.

## License
This package is licensed under the MIT License.