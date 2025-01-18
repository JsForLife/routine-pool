package routinepool

// worker 工作协程结构体
type worker struct {
	id        int
	owner     *Pool
	requestCh chan Task
	shutdown  chan struct{}
}

// start 启动工作协程
func (w *worker) start() {
	defer w.owner.wg.Done()
	for {
		select {
		case task, ok := <-w.requestCh:
			if !ok {
				return
			}
			result, err := task.fn()
			// 将结果发送到调用方提供的响应通道
			if task.responseCh != nil {
				task.responseCh <- Result{
					index: task.index,
					value: result,
					err:   err,
				}
			}
		case <-w.shutdown:
			return
		case <-w.owner.shutdownCh:
			return
		}
	}
}

func (w *worker) stop() {
	close(w.shutdown)
}
