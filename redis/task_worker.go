package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type TaskHandler func(ctx context.Context, task *Task) error

type TaskWorker struct {
	taskQueue   *TaskQueue
	handlers    map[TaskType]TaskHandler
	workerCount int
	running     bool
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

func NewTaskWorker(taskQueue *TaskQueue, workerCount int) *TaskWorker {
	worker := &TaskWorker{
		taskQueue:   taskQueue,
		handlers:    make(map[TaskType]TaskHandler),
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}

	return worker
}

// RegisterHandler ลงทะเบียน handler สำหรับ task type ใหม่
func (tw *TaskWorker) RegisterHandler(taskType TaskType, handler TaskHandler) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.handlers[taskType] = handler
}

// Start เริ่มการทำงานของ worker
func (tw *TaskWorker) Start(ctx context.Context) {
	tw.mu.Lock()
	if tw.running {
		tw.mu.Unlock()
		return
	}
	tw.running = true
	tw.mu.Unlock()

	// เริ่ม recovery goroutine
	tw.wg.Add(1)
	go tw.recoveryLoop(ctx)

	// เริ่ม worker goroutines
	for i := 0; i < tw.workerCount; i++ {
		tw.wg.Add(1)
		go tw.worker(ctx, i)
	}

	log.Println("Starting Task worker successfully!!")
}

// Stop หยุดการทำงานของ worker
func (tw *TaskWorker) Stop() {
	tw.mu.Lock()
	if !tw.running {
		tw.mu.Unlock()
		return
	}
	tw.running = false
	tw.mu.Unlock()

	close(tw.stopChan)
	tw.wg.Wait()

	log.Println("Stopping Task worker successfully!!")
}

// worker function สำหรับประมวลผล tasks
func (tw *TaskWorker) worker(ctx context.Context, workerID int) {
	defer tw.wg.Done()

	for {
		select {
		case <-tw.stopChan:
			log.Printf("Worker %d stopping", workerID)
			return
		case <-ctx.Done():
			log.Printf("Worker %d stopping due to context cancellation", workerID)
			return
		default:
			// Dequeue task with timeout
			task, err := tw.taskQueue.DequeueTask(ctx, 5*time.Second)
			if err != nil {
				log.Printf("Worker %d failed to dequeue task: %v", workerID, err)
				time.Sleep(1 * time.Second)
				continue
			}

			if task == nil {
				// No task available, continue
				continue
			}

			log.Printf("Worker %d processing task %s (type: %s)", workerID, task.ID, task.Type)

			// Process task
			err = tw.processTask(ctx, task)
			if err != nil {
				log.Printf("Worker %d failed to process task %s: %v", workerID, task.ID, err)

				// Mark task as failed/retry
				failErr := tw.taskQueue.FailTask(ctx, task, err.Error())
				if failErr != nil {
					log.Printf("Worker %d failed to mark task %s as failed: %v", workerID, task.ID, failErr)
				}
			} else {
				log.Printf("Worker %d completed task %s successfully", workerID, task.ID)

				// Mark task as completed
				completeErr := tw.taskQueue.CompleteTask(ctx, task)
				if completeErr != nil {
					log.Printf("Worker %d failed to mark task %s as completed: %v", workerID, task.ID, completeErr)
				}
			}
		}
	}
}

// processTask ประมวลผล task ตาม type
func (tw *TaskWorker) processTask(ctx context.Context, task *Task) error {
	tw.mu.RLock()
	handler, exists := tw.handlers[task.Type]
	tw.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for task type: %s", task.Type)
	}

	// Create context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	return handler(taskCtx, task)
}

// recoveryLoop ทำงานเป็นระยะเพื่อ recover stuck tasks
func (tw *TaskWorker) recoveryLoop(ctx context.Context) {
	defer tw.wg.Done()

	ticker := time.NewTicker(2 * time.Minute) // Check every 2 minutes
	defer ticker.Stop()

	for {
		select {
		case <-tw.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Recover tasks that have been stuck for more than 5 minutes
			err := tw.taskQueue.RecoverStuckTasks(ctx, 5*time.Minute)
			if err != nil {
				log.Printf("Failed to recover stuck tasks: %v", err)
			}
		}
	}
}

// GetStats ดูสถิติของ worker
func (tw *TaskWorker) GetStats(ctx context.Context) (map[string]interface{}, error) {
	queueStats, err := tw.taskQueue.GetQueueStats(ctx)
	if err != nil {
		return nil, err
	}

	tw.mu.RLock()
	running := tw.running
	workerCount := tw.workerCount
	handlerCount := len(tw.handlers)
	tw.mu.RUnlock()

	stats := map[string]interface{}{
		"worker_count":  workerCount,
		"handler_count": handlerCount,
		"running":       running,
		"queue_stats":   queueStats,
	}

	return stats, nil
}
