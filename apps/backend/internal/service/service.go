package service

import (
	"context"
	"errors"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

var (
	// ErrEmailAlreadyExists is returned when trying to register with an existing email
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when email or password is invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrAccountLocked is returned when account is temporarily locked
	ErrAccountLocked = errors.New("account is locked")
)

const (
	// MaxFailedLoginAttempts is the maximum number of failed login attempts before account lock
	MaxFailedLoginAttempts = 3
	// AccountLockDuration is the duration for which an account is locked
	AccountLockDuration = 30 * time.Minute
)

// Service provides business logic
type Service struct {
	repos repository.Repositories
}

// NewService creates a new Service instance
func NewService(repos repository.Repositories) *Service {
	return &Service{repos: repos}
}

// --- User Service Methods ---

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, user *model.User) error {
	return s.repos.User().Create(ctx, user)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.repos.User().GetByID(ctx, id)
}

// GetUserByFirebaseUID retrieves a user by Firebase UID
func (s *Service) GetUserByFirebaseUID(ctx context.Context, uid string) (*model.User, error) {
	return s.repos.User().GetByFirebaseUID(ctx, uid)
}

// GetOrCreateUser gets an existing user or creates a new one (with transaction)
func (s *Service) GetOrCreateUser(ctx context.Context, firebaseUID, email, displayName, photoURL string) (*model.User, error) {
	var result *model.User

	err := s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repos.User().GetByFirebaseUID(txCtx, firebaseUID)
		if err == nil {
			result = user
			return nil
		}

		// Create new user
		newUser := &model.User{
			FirebaseUID: firebaseUID,
			Email:       email,
			DisplayName: displayName,
			PhotoURL:    photoURL,
			IsActive:    true,
		}

		if err := s.repos.User().Create(txCtx, newUser); err != nil {
			return err
		}

		result = newUser
		return nil
	})

	return result, err
}

// RegisterUser creates a new user with email and password (with transaction)
func (s *Service) RegisterUser(ctx context.Context, email, hashedPassword, displayName string) (*model.User, error) {
	var result *model.User

	err := s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if email already exists
		existingUser, err := s.repos.User().GetByEmail(txCtx, email)
		if err == nil && existingUser != nil {
			return ErrEmailAlreadyExists
		}

		// Create new user
		newUser := &model.User{
			Email:        email,
			PasswordHash: hashedPassword,
			DisplayName:  displayName,
			IsActive:     true,
		}

		if err := s.repos.User().Create(txCtx, newUser); err != nil {
			return err
		}

		result = newUser
		return nil
	})

	return result, err
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.repos.User().GetByEmail(ctx, email)
}

// IncrementFailedLogin increments failed login count and locks account if needed
func (s *Service) IncrementFailedLogin(ctx context.Context, user *model.User) error {
	user.FailedLoginCount++
	if user.FailedLoginCount >= MaxFailedLoginAttempts {
		lockUntil := time.Now().Add(AccountLockDuration)
		user.LockedUntil = &lockUntil
	}
	return s.repos.User().Update(ctx, user)
}

// ResetFailedLogin resets the failed login count on successful login
func (s *Service) ResetFailedLogin(ctx context.Context, user *model.User) error {
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	return s.repos.User().Update(ctx, user)
}

// IsAccountLocked checks if the account is currently locked
func (s *Service) IsAccountLocked(user *model.User) bool {
	if user.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*user.LockedUntil)
}

// --- Garden Service Methods ---

// CreateGarden creates a new garden for a user
func (s *Service) CreateGarden(ctx context.Context, userID uint, name, description, location string, sizeM2 float64) (*model.Garden, error) {
	garden := &model.Garden{
		UserID:      userID,
		Name:        name,
		Description: description,
		Location:    location,
		SizeM2:      sizeM2,
	}

	if err := s.repos.Garden().Create(ctx, garden); err != nil {
		return nil, err
	}

	return garden, nil
}

// GetGardenByID retrieves a garden by ID
func (s *Service) GetGardenByID(ctx context.Context, id uint) (*model.Garden, error) {
	return s.repos.Garden().GetByID(ctx, id)
}

// GetUserGardens retrieves all gardens for a user
func (s *Service) GetUserGardens(ctx context.Context, userID uint) ([]model.Garden, error) {
	return s.repos.Garden().GetByUserID(ctx, userID)
}

// UpdateGarden updates a garden
func (s *Service) UpdateGarden(ctx context.Context, garden *model.Garden) error {
	return s.repos.Garden().Update(ctx, garden)
}

// DeleteGarden soft deletes a garden and all its plants (with transaction)
func (s *Service) DeleteGarden(ctx context.Context, id uint) error {
	return s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// Batch delete all plants in the garden (prevents N+1 query problem)
		if err := s.repos.Plant().DeleteByGardenID(txCtx, id); err != nil {
			return err
		}

		// Delete the garden
		return s.repos.Garden().Delete(txCtx, id)
	})
}

// --- Plant Service Methods ---

// CreatePlant creates a new plant in a garden
func (s *Service) CreatePlant(ctx context.Context, plant *model.Plant) error {
	return s.repos.Plant().Create(ctx, plant)
}

// GetPlantByID retrieves a plant by ID
func (s *Service) GetPlantByID(ctx context.Context, id uint) (*model.Plant, error) {
	return s.repos.Plant().GetByID(ctx, id)
}

// GetGardenPlants retrieves all plants in a garden
func (s *Service) GetGardenPlants(ctx context.Context, gardenID uint) ([]model.Plant, error) {
	return s.repos.Plant().GetByGardenID(ctx, gardenID)
}

// UpdatePlant updates a plant
func (s *Service) UpdatePlant(ctx context.Context, plant *model.Plant) error {
	return s.repos.Plant().Update(ctx, plant)
}

// DeletePlant soft deletes a plant
func (s *Service) DeletePlant(ctx context.Context, id uint) error {
	return s.repos.Plant().Delete(ctx, id)
}

// --- CareLog Service Methods ---

// CreateCareLog creates a new care log for a plant
func (s *Service) CreateCareLog(ctx context.Context, careLog *model.CareLog) error {
	return s.repos.CareLog().Create(ctx, careLog)
}

// GetPlantCareLogs retrieves all care logs for a plant
func (s *Service) GetPlantCareLogs(ctx context.Context, plantID uint) ([]model.CareLog, error) {
	return s.repos.CareLog().GetByPlantID(ctx, plantID)
}

// --- Token Blacklist Service Methods ---

// BlacklistToken adds a token to the blacklist
func (s *Service) BlacklistToken(ctx context.Context, tokenHash string, expiresAt time.Time) error {
	return s.repos.TokenBlacklist().Add(ctx, tokenHash, expiresAt)
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (s *Service) CleanupExpiredTokens(ctx context.Context) error {
	return s.repos.TokenBlacklist().DeleteExpired(ctx)
}

// =============================================================================
// Task Service Methods - タスク管理サービスメソッド
// =============================================================================
// タスク（やることリスト）のCRUD操作を提供します。
// タスクは植物の世話リマインダーや一般的なガーデニング作業に使用されます。

// CreateTask は新しいタスクを作成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - task: 作成するタスク（UserID, Title, DueDateは必須）
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreateTask(ctx context.Context, task *model.Task) error {
	return s.repos.Task().Create(ctx, task)
}

// GetTaskByID はIDでタスクを取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: タスクID
//
// 戻り値:
//   - *model.Task: 見つかったタスク
//   - error: タスクが見つからない場合は gorm.ErrRecordNotFound
func (s *Service) GetTaskByID(ctx context.Context, id uint) (*model.Task, error) {
	return s.repos.Task().GetByID(ctx, id)
}

// GetUserTasks はユーザーの全タスクを取得します。
// 期限日（DueDate）の昇順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []model.Task: タスクの一覧（期限日順）
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	return s.repos.Task().GetByUserID(ctx, userID)
}

// GetUserTasksByStatus はステータスでフィルタリングしたタスクを取得します。
//
// 有効なステータス:
//   - "pending": 未完了
//   - "completed": 完了済み
//   - "cancelled": キャンセル
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//   - status: フィルタするステータス
//
// 戻り値:
//   - []model.Task: 該当するタスクの一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserTasksByStatus(ctx context.Context, userID uint, status string) ([]model.Task, error) {
	return s.repos.Task().GetByUserIDAndStatus(ctx, userID, status)
}

// GetTodayTasks は今日が期限のタスクを取得します。
// ダッシュボードの「今日のタスク」表示に使用されます。
// 優先度降順、期限日昇順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []model.Task: 今日が期限の未完了タスク
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetTodayTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	return s.repos.Task().GetTodayTasks(ctx, userID)
}

// GetOverdueTasks は期限切れのタスクを取得します。
// ダッシュボードの「期限切れ」アラート表示に使用されます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []model.Task: 期限が過ぎた未完了タスク
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetOverdueTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	return s.repos.Task().GetOverdueTasks(ctx, userID)
}

// UpdateTask はタスクを更新します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - task: 更新するタスク（IDは必須）
//
// 戻り値:
//   - error: 更新に失敗した場合のエラー
func (s *Service) UpdateTask(ctx context.Context, task *model.Task) error {
	return s.repos.Task().Update(ctx, task)
}

// CompleteTask はタスクを完了としてマークします。
// Status を "completed" に、CompletedAt を現在時刻に設定します。
// 繰り返し設定がある場合、次回タスクを自動生成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - taskID: 完了するタスクのID
//
// 戻り値:
//   - error: タスクが見つからない、または更新に失敗した場合のエラー
//
// 繰り返しタスクの自動生成条件:
//   - Recurrence が設定されている（daily, weekly, monthly）
//   - MaxOccurrences に達していない（nilの場合は無制限）
//   - RecurrenceEndDate を過ぎていない（nilの場合は無期限）
func (s *Service) CompleteTask(ctx context.Context, taskID uint) error {
	return s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// まずタスクを取得
		task, err := s.repos.Task().GetByID(txCtx, taskID)
		if err != nil {
			return err
		}

		// 完了状態に更新
		now := time.Now()
		task.Status = "completed"
		task.CompletedAt = &now
		task.OccurrenceCount++

		if err := s.repos.Task().Update(txCtx, task); err != nil {
			return err
		}

		// 繰り返しタスクの場合、次回タスクを生成
		if task.Recurrence != "" {
			return s.generateNextRecurringTask(txCtx, task)
		}

		return nil
	})
}

// generateNextRecurringTask は繰り返しタスクの次回タスクを生成します。
//
// 生成条件:
//   - MaxOccurrences が nil、またはまだ上限に達していない
//   - RecurrenceEndDate が nil、または次回期限日がその日付以前
//
// 次回期限日の計算:
//   - daily: DueDate + (RecurrenceInterval * 日)
//   - weekly: DueDate + (RecurrenceInterval * 週)
//   - monthly: DueDate + (RecurrenceInterval * 月)
func (s *Service) generateNextRecurringTask(ctx context.Context, completedTask *model.Task) error {
	// MaxOccurrences チェック
	if completedTask.MaxOccurrences != nil && completedTask.OccurrenceCount >= *completedTask.MaxOccurrences {
		// 最大回数に達したので生成しない
		return nil
	}

	// 次回期限日を計算
	nextDueDate := s.calculateNextDueDate(completedTask.DueDate, completedTask.Recurrence, completedTask.RecurrenceInterval)

	// RecurrenceEndDate チェック
	if completedTask.RecurrenceEndDate != nil && nextDueDate.After(*completedTask.RecurrenceEndDate) {
		// 終了日を過ぎたので生成しない
		return nil
	}

	// 元タスクのIDを決定（既に子タスクの場合は元のParentTaskIDを使用）
	var parentID uint
	if completedTask.ParentTaskID != nil {
		parentID = *completedTask.ParentTaskID
	} else {
		parentID = completedTask.ID
	}

	// 新しいタスクを作成
	newTask := &model.Task{
		UserID:             completedTask.UserID,
		PlantID:            completedTask.PlantID,
		Title:              completedTask.Title,
		Description:        completedTask.Description,
		DueDate:            nextDueDate,
		Priority:           completedTask.Priority,
		Status:             "pending",
		Recurrence:         completedTask.Recurrence,
		RecurrenceInterval: completedTask.RecurrenceInterval,
		MaxOccurrences:     completedTask.MaxOccurrences,
		RecurrenceEndDate:  completedTask.RecurrenceEndDate,
		OccurrenceCount:    completedTask.OccurrenceCount,
		ParentTaskID:       &parentID,
	}

	return s.repos.Task().Create(ctx, newTask)
}

// calculateNextDueDate は次回の期限日を計算します。
//
// 引数:
//   - currentDueDate: 現在の期限日
//   - recurrence: 繰り返し頻度（daily, weekly, monthly）
//   - interval: 間隔
//
// 戻り値:
//   - time.Time: 次回の期限日
func (s *Service) calculateNextDueDate(currentDueDate time.Time, recurrence string, interval int) time.Time {
	if interval <= 0 {
		interval = 1
	}

	switch recurrence {
	case "daily":
		return currentDueDate.AddDate(0, 0, interval)
	case "weekly":
		return currentDueDate.AddDate(0, 0, interval*7)
	case "monthly":
		return currentDueDate.AddDate(0, interval, 0)
	default:
		// 不明な繰り返し頻度の場合は1日後
		return currentDueDate.AddDate(0, 0, 1)
	}
}

// DeleteTask はタスクを論理削除します。
// GORMのソフトデリートにより、DeletedAtが設定されます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: 削除するタスクのID
//
// 戻り値:
//   - error: 削除に失敗した場合のエラー
func (s *Service) DeleteTask(ctx context.Context, id uint) error {
	return s.repos.Task().Delete(ctx, id)
}
