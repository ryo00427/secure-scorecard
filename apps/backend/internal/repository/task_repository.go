package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// taskRepository implements TaskRepository
type taskRepository struct {
	db *gorm.DB
}

// Create creates a new task
func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	return GetDB(ctx, r.db).Create(task).Error
}

// GetByID retrieves a task by ID
func (r *taskRepository) GetByID(ctx context.Context, id uint) (*model.Task, error) {
	var task model.Task
	if err := GetDB(ctx, r.db).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUserID retrieves all tasks for a user
func (r *taskRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Task, error) {
	var tasks []model.Task
	if err := GetDB(ctx, r.db).Where("user_id = ?", userID).Order("due_date ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByUserIDAndStatus retrieves tasks for a user with a specific status
func (r *taskRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Task, error) {
	var tasks []model.Task
	if err := GetDB(ctx, r.db).Where("user_id = ? AND status = ?", userID, status).Order("due_date ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTodayTasks retrieves tasks due today for a user
func (r *taskRepository) GetTodayTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	var tasks []model.Task
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	if err := GetDB(ctx, r.db).
		Where("user_id = ? AND status = ? AND due_date >= ? AND due_date < ?",
			userID, "pending", today, tomorrow).
		Order("priority DESC, due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetOverdueTasks retrieves overdue tasks for a user
func (r *taskRepository) GetOverdueTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	var tasks []model.Task
	today := time.Now().Truncate(24 * time.Hour)

	if err := GetDB(ctx, r.db).
		Where("user_id = ? AND status = ? AND due_date < ?",
			userID, "pending", today).
		Order("due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetAllOverdueTasks はシステム全体の期限切れタスクを取得します（通知処理用）
// ユーザー情報を含めて取得し、通知対象の判定に使用します
func (r *taskRepository) GetAllOverdueTasks(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	today := time.Now().Truncate(24 * time.Hour)

	if err := GetDB(ctx, r.db).
		Preload("User").
		Where("status = ? AND due_date < ?", "pending", today).
		Order("user_id ASC, due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetAllTodayTasks はシステム全体の今日が期限のタスクを取得します（通知処理用）
// ユーザー情報を含めて取得し、リマインダー通知に使用します
func (r *taskRepository) GetAllTodayTasks(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	if err := GetDB(ctx, r.db).
		Preload("User").
		Where("status = ? AND due_date >= ? AND due_date < ?", "pending", today, tomorrow).
		Order("user_id ASC, priority DESC, due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// Update updates a task
func (r *taskRepository) Update(ctx context.Context, task *model.Task) error {
	return GetDB(ctx, r.db).Save(task).Error
}

// Delete soft deletes a task
func (r *taskRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.Task{}, id).Error
}
