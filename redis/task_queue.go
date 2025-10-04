package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	rdb "github.com/redis/go-redis/v9"
)

const (
	TaskQueueKey      = "task_queue"
	TaskProcessingKey = "task_processing"
	TaskRetryKey      = "task_retry"
	TaskFailedKey     = "task_failed"
	DefaultRetryLimit = 3
	DefaultTimeout    = 30 * time.Second
)

type TaskType string

// NewTaskType สร้าง TaskType ใหม่
func NewTaskType(name string) TaskType {
	return TaskType(name)
}

// String returns the string representation of TaskType
func (t TaskType) String() string {
	return string(t)
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRetrying   TaskStatus = "retrying"
)

type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Status      TaskStatus             `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

type TaskQueue struct {
	client *Client
}

func NewTaskQueue(client *Client) *TaskQueue {
	return &TaskQueue{
		client: client,
	}
}

// EnqueueTask เพิ่ม task ใหม่เข้า queue
func (tq *TaskQueue) EnqueueTask(ctx context.Context, taskType TaskType, payload map[string]interface{}) (*Task, error) {
	taskID, _ := uuid.NewV4()
	task := &Task{
		ID:         taskID.String(),
		Type:       taskType,
		Status:     TaskStatusPending,
		Payload:    payload,
		RetryCount: 0,
		MaxRetries: DefaultRetryLimit,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %v", err)
	}

	// เพิ่ม task เข้า queue
	err = tq.client.rdbc.LPush(ctx, TaskQueueKey, taskJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %v", err)
	}

	// เก็บ task detail ใน hash
	err = tq.client.rdbc.HSet(ctx, fmt.Sprintf("task:%s", task.ID), task.ID, taskJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store task details: %v", err)
	}

	log.Printf("Task %s enqueued successfully", task.ID)
	return task, nil
}

// DequeueTask ดึง task จาก queue มาประมวลผล
func (tq *TaskQueue) DequeueTask(ctx context.Context, timeout time.Duration) (*Task, error) {
	// ใช้ BRPOPLPUSH เพื่อย้าย task จาก queue ไป processing list อย่างปลอดภัย
	result := tq.client.rdbc.BRPopLPush(ctx, TaskQueueKey, TaskProcessingKey, timeout)
	if result.Err() != nil {
		if result.Err() == rdb.Nil {
			return nil, nil // ไม่มี task
		}
		return nil, fmt.Errorf("failed to dequeue task: %v", result.Err())
	}

	var task Task
	err := json.Unmarshal([]byte(result.Val()), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %v", err)
	}

	// อัพเดทสถานะเป็น processing
	task.Status = TaskStatusProcessing
	task.UpdatedAt = time.Now()
	now := time.Now()
	task.ProcessedAt = &now

	err = tq.updateTaskStatus(ctx, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %v", err)
	}

	return &task, nil
}

// CompleteTask ทำเครื่องหมายว่า task เสร็จสิ้นแล้ว
func (tq *TaskQueue) CompleteTask(ctx context.Context, task *Task) error {
	task.Status = TaskStatusCompleted
	task.UpdatedAt = time.Now()

	// ลบจาก processing list
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	err = tq.client.rdbc.LRem(ctx, TaskProcessingKey, 1, taskJSON).Err()
	if err != nil {
		log.Printf("Warning: failed to remove task from processing list: %v", err)
	}

	// อัพเดทสถานะ
	err = tq.updateTaskStatus(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	// ลบ task detail หลังจากเสร็จสิ้น
	err = tq.client.rdbc.Del(ctx, fmt.Sprintf("task:%s", task.ID)).Err()
	if err != nil {
		log.Printf("Warning: failed to delete task details: %v", err)
	}

	return nil
}

// FailTask ทำเครื่องหมายว่า task ล้มเหลวหรือ retry
func (tq *TaskQueue) FailTask(ctx context.Context, task *Task, errorMsg string) error {
	task.RetryCount++
	task.UpdatedAt = time.Now()
	task.ErrorMsg = errorMsg

	if task.RetryCount >= task.MaxRetries {
		// Task ล้มเหลวสุดท้าย
		task.Status = TaskStatusFailed
		now := time.Now()
		task.FailedAt = &now

		// ย้ายไป failed queue
		taskJSON, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task: %v", err)
		}

		err = tq.client.rdbc.LPush(ctx, TaskFailedKey, taskJSON).Err()
		if err != nil {
			log.Printf("Warning: failed to add task to failed queue: %v", err)
		}

		log.Printf("Task %s failed permanently after %d retries: %s", task.ID, task.RetryCount, errorMsg)
	} else {
		// Retry task
		task.Status = TaskStatusRetrying

		// คำนวณ delay สำหรับ retry (exponential backoff)
		delay := time.Duration(task.RetryCount*task.RetryCount) * time.Second
		if delay > 5*time.Minute {
			delay = 5 * time.Minute
		}

		// เพิ่มกลับเข้า queue หลังจาก delay
		taskJSON, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task: %v", err)
		}

		// ใช้ delayed queue pattern
		go func() {
			time.Sleep(delay)
			err := tq.client.rdbc.LPush(context.Background(), TaskQueueKey, taskJSON).Err()
			if err != nil {
				log.Printf("Failed to re-enqueue task %s: %v", task.ID, err)
			} else {
				log.Printf("Task %s re-enqueued for retry %d/%d after %v delay",
					task.ID, task.RetryCount, task.MaxRetries, delay)
			}
		}()
	}

	// ลบจาก processing list
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	err = tq.client.rdbc.LRem(ctx, TaskProcessingKey, 1, taskJSON).Err()
	if err != nil {
		log.Printf("Warning: failed to remove task from processing list: %v", err)
	}

	// อัพเดทสถานะ
	err = tq.updateTaskStatus(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	return nil
}

// GetTaskStatus ดูสถานะของ task
func (tq *TaskQueue) GetTaskStatus(ctx context.Context, taskID string) (*Task, error) {
	result := tq.client.rdbc.HGet(ctx, fmt.Sprintf("task:%s", taskID), taskID)
	if result.Err() != nil {
		if result.Err() == rdb.Nil {
			return nil, nil // ไม่พบ task
		}
		return nil, fmt.Errorf("failed to get task status: %v", result.Err())
	}

	var task Task
	err := json.Unmarshal([]byte(result.Val()), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %v", err)
	}

	return &task, nil
}

// RecoverStuckTasks ดึง tasks ที่ค้างอยู่ใน processing กลับมา queue
func (tq *TaskQueue) RecoverStuckTasks(ctx context.Context, stuckTimeout time.Duration) error {
	processingTasks := tq.client.rdbc.LRange(ctx, TaskProcessingKey, 0, -1)
	if processingTasks.Err() != nil {
		return fmt.Errorf("failed to get processing tasks: %v", processingTasks.Err())
	}

	recoveredCount := 0
	for _, taskJSON := range processingTasks.Val() {
		var task Task
		err := json.Unmarshal([]byte(taskJSON), &task)
		if err != nil {
			log.Printf("Failed to unmarshal processing task: %v", err)
			continue
		}

		// ตรวจสอบว่า task ค้างนานเกินกำหนดหรือไม่
		if task.ProcessedAt != nil && time.Since(*task.ProcessedAt) > stuckTimeout {
			// ย้ายกลับเข้า queue
			err = tq.client.rdbc.LRem(ctx, TaskProcessingKey, 1, taskJSON).Err()
			if err != nil {
				log.Printf("Failed to remove stuck task from processing: %v", err)
				continue
			}

			task.Status = TaskStatusPending
			task.UpdatedAt = time.Now()
			task.ProcessedAt = nil

			newTaskJSON, err := json.Marshal(task)
			if err != nil {
				log.Printf("Failed to marshal recovered task: %v", err)
				continue
			}

			err = tq.client.rdbc.LPush(ctx, TaskQueueKey, newTaskJSON).Err()
			if err != nil {
				log.Printf("Failed to re-enqueue recovered task: %v", err)
				continue
			}

			err = tq.updateTaskStatus(ctx, &task)
			if err != nil {
				log.Printf("Failed to update recovered task status: %v", err)
			}

			recoveredCount++
			log.Printf("Recovered stuck task %s", task.ID)
		}
	}

	if recoveredCount > 0 {
		log.Printf("Recovered %d stuck tasks", recoveredCount)
	}

	return nil
}

// GetQueueStats ดูสถิติของ queue
func (tq *TaskQueue) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// นับ pending tasks
	pendingCount := tq.client.rdbc.LLen(ctx, TaskQueueKey)
	if pendingCount.Err() == nil {
		stats["pending"] = pendingCount.Val()
	}

	// นับ processing tasks
	processingCount := tq.client.rdbc.LLen(ctx, TaskProcessingKey)
	if processingCount.Err() == nil {
		stats["processing"] = processingCount.Val()
	}

	// นับ failed tasks
	failedCount := tq.client.rdbc.LLen(ctx, TaskFailedKey)
	if failedCount.Err() == nil {
		stats["failed"] = failedCount.Val()
	}

	return stats, nil
}

// updateTaskStatus อัพเดทสถานะของ task ใน Redis
func (tq *TaskQueue) updateTaskStatus(ctx context.Context, task *Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	err = tq.client.rdbc.HSet(ctx, fmt.Sprintf("task:%s", task.ID), task.ID, taskJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	return nil
}
