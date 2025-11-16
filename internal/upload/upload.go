package upload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// UploadStatus 上传状态
type UploadStatus string

const (
	UploadStatusUploading UploadStatus = "uploading" // 上传中
	UploadStatusMerged    UploadStatus = "merged"    // 已合并
	UploadStatusFailed    UploadStatus = "failed"    // 失败
)

// UploadTask 上传任务
type UploadTask struct {
	UploadID       string       `json:"uploadId"`
	FileName       string       `json:"fileName"`
	FileSize       int64        `json:"fileSize"`
	TotalChunks    int          `json:"totalChunks"`
	ChunkSize      int64        `json:"chunkSize"`
	UploadedChunks int          `json:"uploadedChunks"`
	Status         UploadStatus `json:"status"`
	MergedPath     string       `json:"mergedPath,omitempty"`
	TempDir        string       `json:"-"` // 临时目录,不序列化
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	ctx            context.Context
	cancel         context.CancelFunc
	chunks         map[int]bool // 已上传的切片索引
}

// Manager 上传管理器
type Manager struct {
	tasks   map[string]*UploadTask
	mu      sync.RWMutex
	tempDir string // 临时文件目录
	dataDir string // 数据目录
}

// NewManager 创建上传管理器
func NewManager(tempDir, dataDir string) *Manager {
	return &Manager{
		tasks:   make(map[string]*UploadTask),
		tempDir: tempDir,
		dataDir: dataDir,
	}
}

// CreateUploadTask 创建上传任务
func (m *Manager) CreateUploadTask(fileName string, fileSize int64, totalChunks int, chunkSize int64) (*UploadTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	uploadID := uuid.New().String()

	// 创建临时目录
	tempDir := filepath.Join(m.tempDir, uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}

	task := &UploadTask{
		UploadID:       uploadID,
		FileName:       fileName,
		FileSize:       fileSize,
		TotalChunks:    totalChunks,
		ChunkSize:      chunkSize,
		UploadedChunks: 0,
		Status:         UploadStatusUploading,
		TempDir:        tempDir,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ctx:            ctx,
		cancel:         cancel,
		chunks:         make(map[int]bool),
	}

	m.tasks[uploadID] = task
	return task, nil
}

// RecordChunk 记录切片上传
func (m *Manager) RecordChunk(uploadID string, chunkIndex int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[uploadID]
	if !ok {
		return fmt.Errorf("上传任务不存在: %s", uploadID)
	}

	if !task.chunks[chunkIndex] {
		task.chunks[chunkIndex] = true
		task.UploadedChunks++
		task.UpdatedAt = time.Now()
	}

	return nil
}

// GetUploadTask 获取上传任务
func (m *Manager) GetUploadTask(uploadID string) (*UploadTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[uploadID]
	if !ok {
		return nil, fmt.Errorf("上传任务不存在: %s", uploadID)
	}
	return task, nil
}

// MergeChunks 合并切片
func (m *Manager) MergeChunks(uploadID string) error {
	task, err := m.GetUploadTask(uploadID)
	if err != nil {
		return err
	}

	// 输出文件路径
	mergedPath := filepath.Join(m.dataDir, uploadID+"_"+task.FileName)

	// 创建输出文件
	outFile, err := os.Create(mergedPath)
	if err != nil {
		return fmt.Errorf("创建合并文件失败: %v", err)
	}
	defer outFile.Close()

	// 按顺序合并切片
	for i := 0; i < task.TotalChunks; i++ {
		chunkPath := filepath.Join(task.TempDir, fmt.Sprintf("chunk_%d", i))

		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("读取切片 %d 失败: %v", i, err)
		}

		if _, err := outFile.Write(chunkData); err != nil {
			return fmt.Errorf("写入切片 %d 失败: %v", i, err)
		}
	}

	// 更新任务状态
	m.mu.Lock()
	task.Status = UploadStatusMerged
	task.MergedPath = mergedPath
	task.UpdatedAt = time.Now()
	m.mu.Unlock()

	// 清理临时目录
	go func() {
		time.Sleep(1 * time.Second)
		os.RemoveAll(task.TempDir)
	}()

	return nil
}

// CancelUpload 取消上传
func (m *Manager) CancelUpload(uploadID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[uploadID]
	if !ok {
		return fmt.Errorf("上传任务不存在: %s", uploadID)
	}

	if task.cancel != nil {
		task.cancel()
	}

	// 清理临时文件
	if task.TempDir != "" {
		os.RemoveAll(task.TempDir)
	}

	delete(m.tasks, uploadID)
	return nil
}

// GetChunkPath 获取切片文件路径
func (t *UploadTask) GetChunkPath(chunkIndex int) string {
	return filepath.Join(t.TempDir, fmt.Sprintf("chunk_%d", chunkIndex))
}

// IsComplete 检查是否所有切片都已上传
func (t *UploadTask) IsComplete() bool {
	return t.UploadedChunks == t.TotalChunks
}
