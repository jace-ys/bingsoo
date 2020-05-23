package worker

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrMaxCapacity = errors.New("task queue is at max capacity")
)

type Task interface {
	Process(ctx context.Context) error
}

type WorkerPool struct {
	taskChan  chan Task
	waitGroup sync.WaitGroup
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{}
}

func (p *WorkerPool) Process(ctx context.Context, concurrency int) error {
	p.taskChan = make(chan Task, concurrency)

	p.waitGroup.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer p.waitGroup.Done()
			for task := range p.taskChan {
				task.Process(ctx)
			}
		}()
	}

	p.waitGroup.Wait()
	return nil
}

func (p *WorkerPool) Enqueue(task Task) error {
	select {
	case p.taskChan <- task:
		return nil
	default:
		return ErrMaxCapacity
	}
}

func (p *WorkerPool) Close() error {
	close(p.taskChan)
	return nil
}
