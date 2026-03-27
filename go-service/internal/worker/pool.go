package worker

import (
	"context"
	"log"
	"sync"

	"agro-subsidy/go-service/internal/model"
	"agro-subsidy/go-service/internal/service"
)

// Pool runs N goroutines that pull SubsidiesTasks from an internal queue,
// call the ML service, and push MLResponses to a results channel.
// The bounded queue (size = QueueSize) is the back-pressure mechanism:
// if ML is slow the Kafka consumer will block instead of spawning more goroutines.
type Pool struct {
	workerCount int
	queue       chan model.SubsidiesTask
	results     chan model.MLResponse
	mlClient    *service.MLClient
	wg          sync.WaitGroup
}

func NewPool(workerCount, queueSize int, mlClient *service.MLClient) *Pool {
	return &Pool{
		workerCount: workerCount,
		queue:       make(chan model.SubsidiesTask, queueSize),
		results:     make(chan model.MLResponse, queueSize),
		mlClient:    mlClient,
	}
}

// Start launches workerCount goroutines. They run until ctx is cancelled
// or the queue channel is closed (via Stop).
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.queue:
			if !ok {
				return
			}
			req := task.ToMLRequest()
			resp, err := p.mlClient.Score(ctx, req)
			if err != nil {
				log.Printf("[worker %d] score error task=%s: %v", id, task.ID, err)
				continue
			}
			select {
			case p.results <- *resp:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Submit enqueues a task. Blocks if the queue is full (back-pressure).
func (p *Pool) Submit(ctx context.Context, task model.SubsidiesTask) bool {
	select {
	case p.queue <- task:
		return true
	case <-ctx.Done():
		return false
	}
}

// Results returns the read-only channel of scored responses.
func (p *Pool) Results() <-chan model.MLResponse {
	return p.results
}

// Stop closes the task queue, waits for all workers to drain, then closes results.
func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
	close(p.results)
}
