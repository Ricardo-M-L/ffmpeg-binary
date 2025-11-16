package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Status 任务状态
type Status string

const (
	StatusPending    Status = "pending"    // 等待中
	StatusProcessing Status = "processing" // 处理中
	StatusCompleted  Status = "completed"  // 已完成
	StatusFailed     Status = "failed"     // 失败
)

// Task 转换任务
type Task struct {
	ID           string     `json:"taskId"`                // 任务ID
	Status       Status     `json:"status"`                // 状态
	Progress     int        `json:"progress"`              // 进度 0-100
	InputPath    string     `json:"inputPath"`             // 输入文件路径
	OutputPath   string     `json:"outputPath"`            // 输出文件路径
	OutputFormat string     `json:"outputFormat"`          // 输出格式
	Quality      string     `json:"quality"`               // 质量
	UploadID     string     `json:"uploadId,omitempty"`    // 关联的上传ID
	Error        string     `json:"error,omitempty"`       // 错误信息
	CreatedAt    time.Time  `json:"createdAt"`             // 创建时间
	UpdatedAt    time.Time  `json:"updatedAt"`             // 更新时间
	CompletedAt  *time.Time `json:"completedAt,omitempty"` // 完成时间
	ctx          context.Context
	cancel       context.CancelFunc
}

// Manager 任务管理器
type Manager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

// NewManager 创建任务管理器
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

// Create 创建新任务
func (m *Manager) Create(inputPath, outputPath string) *Task {
	return m.CreateWithOptions(inputPath, outputPath, "mp4", "medium", "")
}

// CreateWithOptions 创建新任务(带完整选项)
func (m *Manager) CreateWithOptions(inputPath, outputPath, outputFormat, quality, uploadID string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		ID:           uuid.New().String(),
		Status:       StatusPending,
		Progress:     0,
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: outputFormat,
		Quality:      quality,
		UploadID:     uploadID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}

	m.tasks[task.ID] = task
	return task
}

// MarkCompleted 标记任务完成
func (m *Manager) MarkCompleted(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("任务不存在: %s", id)
	}

	now := time.Now()
	task.Status = StatusCompleted
	task.Progress = 100
	task.CompletedAt = &now
	task.UpdatedAt = now
	return nil
}

// Get 获取任务
func (m *Manager) Get(id string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("任务不存在: %s", id)
	}
	return task, nil
}

// UpdateStatus 更新任务状态
func (m *Manager) UpdateStatus(id string, status Status, progress int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("任务不存在: %s", id)
	}

	task.Status = status
	task.Progress = progress
	task.UpdatedAt = time.Now()
	return nil
}

// UpdateError 更新任务错误信息
func (m *Manager) UpdateError(id string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("任务不存在: %s", id)
	}

	task.Status = StatusFailed
	task.Error = err.Error()
	task.UpdatedAt = time.Now()
	return nil
}

// Delete 删除任务
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if ok && task.cancel != nil {
		task.cancel()
	}

	delete(m.tasks, id)
	return nil
}

// List 列出所有任务
func (m *Manager) List() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// Context 获取任务的上下文
func (t *Task) Context() context.Context {
	return t.ctx
}
