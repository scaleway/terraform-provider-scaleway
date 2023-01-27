package internal

import (
	"sync"
)

type WorkerPoolTask func() error

type WorkerPool struct {
	tasksToDispatch   chan WorkerPoolTask
	tasksToRun        chan WorkerPoolTask
	errors            []error
	errorsMutex       sync.Mutex
	tasksWaitingGroup sync.WaitGroup
}

func NewWorkerPool(size int) *WorkerPool {
	p := &WorkerPool{
		tasksToDispatch: make(chan WorkerPoolTask),
		tasksToRun:      make(chan WorkerPoolTask, size),
	}

	for i := 0; i < size; i++ {
		go p.worker()
	}

	go p.dispatcher()

	return p
}

func (p *WorkerPool) dispatcher() {
	var pendingTasks []WorkerPoolTask

	for {
		if len(pendingTasks) > 0 {
			select {
			case p.tasksToRun <- pendingTasks[0]:
				pendingTasks = pendingTasks[1:]
			default:
			}
		}

		select {
		case task, ok := <-p.tasksToDispatch:
			if !ok {
				if len(pendingTasks) > 0 {
					continue
				}

				close(p.tasksToRun)
				return
			}

			select {
			case p.tasksToRun <- task:
			default:
				pendingTasks = append(pendingTasks, task)
			}
		default:
		}
	}
}

func (p *WorkerPool) worker() {
	for task := range p.tasksToRun {
		err := task()
		if err != nil {
			p.errorsMutex.Lock()
			p.errors = append(p.errors, err)
			p.errorsMutex.Unlock()
		}

		p.tasksWaitingGroup.Done()
	}
}

func (p *WorkerPool) AddTask(task WorkerPoolTask) {
	p.tasksWaitingGroup.Add(1)
	p.tasksToDispatch <- task
}

func (p *WorkerPool) CloseAndWait() []error {
	close(p.tasksToDispatch)
	p.tasksWaitingGroup.Wait()

	return p.errors
}
