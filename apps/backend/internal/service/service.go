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

// =============================================================================
// Crop Service Methods - 作物管理サービスメソッド
// =============================================================================
// 作物（Crop）の植え付けから収穫までのライフサイクルを管理します。
// 成長記録（GrowthRecord）と収穫記録（Harvest）も含みます。

// CreateCrop は新しい作物を登録します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - crop: 作成する作物（UserID, Name, PlantedDate, ExpectedHarvestDateは必須）
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreateCrop(ctx context.Context, crop *model.Crop) error {
	return s.repos.Crop().Create(ctx, crop)
}

// GetCropByID はIDで作物を取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: 作物ID
//
// 戻り値:
//   - *model.Crop: 見つかった作物
//   - error: 作物が見つからない場合は gorm.ErrRecordNotFound
func (s *Service) GetCropByID(ctx context.Context, id uint) (*model.Crop, error) {
	return s.repos.Crop().GetByID(ctx, id)
}

// GetUserCrops はユーザーの全作物を取得します。
// 植え付け日（PlantedDate）の降順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []model.Crop: 作物の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserCrops(ctx context.Context, userID uint) ([]model.Crop, error) {
	return s.repos.Crop().GetByUserID(ctx, userID)
}

// GetUserCropsByStatus はステータスでフィルタリングした作物を取得します。
//
// 有効なステータス:
//   - "planted": 植え付け済み
//   - "growing": 成長中
//   - "ready_to_harvest": 収穫可能
//   - "harvested": 収穫済み
//   - "failed": 失敗
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//   - status: フィルタするステータス
//
// 戻り値:
//   - []model.Crop: 該当する作物の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserCropsByStatus(ctx context.Context, userID uint, status string) ([]model.Crop, error) {
	return s.repos.Crop().GetByUserIDAndStatus(ctx, userID, status)
}

// UpdateCrop は作物を更新します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - crop: 更新する作物（IDは必須）
//
// 戻り値:
//   - error: 更新に失敗した場合のエラー
func (s *Service) UpdateCrop(ctx context.Context, crop *model.Crop) error {
	return s.repos.Crop().Update(ctx, crop)
}

// DeleteCrop は作物と関連する成長記録・収穫記録を削除します（トランザクション使用）。
// N+1問題を避けるため、バッチ削除を使用します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: 削除する作物のID
//
// 戻り値:
//   - error: 削除に失敗した場合のエラー
func (s *Service) DeleteCrop(ctx context.Context, id uint) error {
	return s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// 関連する成長記録を一括削除
		if err := s.repos.GrowthRecord().DeleteByCropID(txCtx, id); err != nil {
			return err
		}

		// 関連する収穫記録を一括削除
		if err := s.repos.Harvest().DeleteByCropID(txCtx, id); err != nil {
			return err
		}

		// 作物を削除
		return s.repos.Crop().Delete(txCtx, id)
	})
}

// =============================================================================
// GrowthRecord Service Methods - 成長記録サービスメソッド
// =============================================================================

// CreateGrowthRecord は新しい成長記録を作成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - record: 作成する成長記録（CropID, RecordDate, GrowthStageは必須）
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreateGrowthRecord(ctx context.Context, record *model.GrowthRecord) error {
	return s.repos.GrowthRecord().Create(ctx, record)
}

// GetGrowthRecordByID はIDで成長記録を取得します。
func (s *Service) GetGrowthRecordByID(ctx context.Context, id uint) (*model.GrowthRecord, error) {
	return s.repos.GrowthRecord().GetByID(ctx, id)
}

// GetCropGrowthRecords は作物の全成長記録を取得します。
// 記録日（RecordDate）の降順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - cropID: 作物ID
//
// 戻り値:
//   - []model.GrowthRecord: 成長記録の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetCropGrowthRecords(ctx context.Context, cropID uint) ([]model.GrowthRecord, error) {
	return s.repos.GrowthRecord().GetByCropID(ctx, cropID)
}

// DeleteGrowthRecord は成長記録を削除します。
func (s *Service) DeleteGrowthRecord(ctx context.Context, id uint) error {
	return s.repos.GrowthRecord().Delete(ctx, id)
}

// =============================================================================
// Harvest Service Methods - 収穫記録サービスメソッド
// =============================================================================

// CreateHarvest は新しい収穫記録を作成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - harvest: 作成する収穫記録（CropID, HarvestDate, Quantity, QuantityUnitは必須）
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreateHarvest(ctx context.Context, harvest *model.Harvest) error {
	return s.repos.Harvest().Create(ctx, harvest)
}

// GetHarvestByID はIDで収穫記録を取得します。
func (s *Service) GetHarvestByID(ctx context.Context, id uint) (*model.Harvest, error) {
	return s.repos.Harvest().GetByID(ctx, id)
}

// GetCropHarvests は作物の全収穫記録を取得します。
// 収穫日（HarvestDate）の降順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - cropID: 作物ID
//
// 戻り値:
//   - []model.Harvest: 収穫記録の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetCropHarvests(ctx context.Context, cropID uint) ([]model.Harvest, error) {
	return s.repos.Harvest().GetByCropID(ctx, cropID)
}

// DeleteHarvest は収穫記録を削除します。
func (s *Service) DeleteHarvest(ctx context.Context, id uint) error {
	return s.repos.Harvest().Delete(ctx, id)
}

// =============================================================================
// Plot Service Methods - 区画管理サービスメソッド
// =============================================================================
// 菜園の区画（Plot）を管理します。
// 区画は作物の配置場所として使用され、グリッドレイアウトをサポートします。

// CreatePlot は新しい区画を作成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plot: 作成する区画（UserID, Name, Width, Heightは必須）
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreatePlot(ctx context.Context, plot *model.Plot) error {
	return s.repos.Plot().Create(ctx, plot)
}

// GetPlotByID はIDで区画を取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: 区画ID
//
// 戻り値:
//   - *model.Plot: 見つかった区画
//   - error: 区画が見つからない場合は gorm.ErrRecordNotFound
func (s *Service) GetPlotByID(ctx context.Context, id uint) (*model.Plot, error) {
	return s.repos.Plot().GetByID(ctx, id)
}

// GetUserPlots はユーザーの全区画を取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []model.Plot: 区画の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserPlots(ctx context.Context, userID uint) ([]model.Plot, error) {
	return s.repos.Plot().GetByUserID(ctx, userID)
}

// GetUserPlotsByStatus はステータスでフィルタリングした区画を取得します。
//
// 有効なステータス:
//   - "available": 空き
//   - "occupied": 使用中
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//   - status: フィルタするステータス
//
// 戻り値:
//   - []model.Plot: 該当する区画の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetUserPlotsByStatus(ctx context.Context, userID uint, status string) ([]model.Plot, error) {
	return s.repos.Plot().GetByUserIDAndStatus(ctx, userID, status)
}

// UpdatePlot は区画を更新します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plot: 更新する区画（IDは必須）
//
// 戻り値:
//   - error: 更新に失敗した場合のエラー
func (s *Service) UpdatePlot(ctx context.Context, plot *model.Plot) error {
	return s.repos.Plot().Update(ctx, plot)
}

// DeletePlot は区画と関連する配置履歴を削除します（トランザクション使用）。
// N+1問題を避けるため、バッチ削除を使用します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - id: 削除する区画のID
//
// 戻り値:
//   - error: 削除に失敗した場合のエラー
func (s *Service) DeletePlot(ctx context.Context, id uint) error {
	return s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// 関連する配置履歴を一括削除
		if err := s.repos.PlotAssignment().DeleteByPlotID(txCtx, id); err != nil {
			return err
		}

		// 区画を削除
		return s.repos.Plot().Delete(txCtx, id)
	})
}

// =============================================================================
// PlotAssignment Service Methods - 区画配置サービスメソッド
// =============================================================================
// 区画への作物配置を管理します。
// 配置履歴を追跡し、過去の配置も記録します。

// AssignCropToPlot は作物を区画に配置します。
// 既存のアクティブな配置がある場合は、まずそれを解除します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plotID: 配置先の区画ID
//   - cropID: 配置する作物ID
//   - assignedDate: 配置日
//
// 戻り値:
//   - *model.PlotAssignment: 作成された配置
//   - error: 配置に失敗した場合のエラー
func (s *Service) AssignCropToPlot(ctx context.Context, plotID, cropID uint, assignedDate time.Time) (*model.PlotAssignment, error) {
	var result *model.PlotAssignment

	err := s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// 既存のアクティブな配置を解除
		existingAssignment, err := s.repos.PlotAssignment().GetActiveByPlotID(txCtx, plotID)
		if err == nil && existingAssignment != nil {
			now := time.Now()
			existingAssignment.UnassignedDate = &now
			if err := s.repos.PlotAssignment().Update(txCtx, existingAssignment); err != nil {
				return err
			}
		}

		// 新しい配置を作成
		assignment := &model.PlotAssignment{
			PlotID:       plotID,
			CropID:       cropID,
			AssignedDate: assignedDate,
		}

		if err := s.repos.PlotAssignment().Create(txCtx, assignment); err != nil {
			return err
		}

		// 区画のステータスを occupied に更新
		plot, err := s.repos.Plot().GetByID(txCtx, plotID)
		if err != nil {
			return err
		}
		plot.Status = "occupied"
		if err := s.repos.Plot().Update(txCtx, plot); err != nil {
			return err
		}

		result = assignment
		return nil
	})

	return result, err
}

// UnassignCropFromPlot は区画から作物の配置を解除します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plotID: 解除する区画ID
//
// 戻り値:
//   - error: 解除に失敗した場合のエラー
func (s *Service) UnassignCropFromPlot(ctx context.Context, plotID uint) error {
	return s.repos.WithTransaction(ctx, func(txCtx context.Context) error {
		// アクティブな配置を取得
		assignment, err := s.repos.PlotAssignment().GetActiveByPlotID(txCtx, plotID)
		if err != nil {
			return err
		}

		// 配置を解除
		now := time.Now()
		assignment.UnassignedDate = &now
		if err := s.repos.PlotAssignment().Update(txCtx, assignment); err != nil {
			return err
		}

		// 区画のステータスを available に更新
		plot, err := s.repos.Plot().GetByID(txCtx, plotID)
		if err != nil {
			return err
		}
		plot.Status = "available"
		return s.repos.Plot().Update(txCtx, plot)
	})
}

// GetPlotAssignments は区画の全配置履歴を取得します。
// 配置日（AssignedDate）の降順でソートされます。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plotID: 区画ID
//
// 戻り値:
//   - []model.PlotAssignment: 配置履歴の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetPlotAssignments(ctx context.Context, plotID uint) ([]model.PlotAssignment, error) {
	return s.repos.PlotAssignment().GetByPlotID(ctx, plotID)
}

// GetActivePlotAssignment は区画の現在アクティブな配置を取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plotID: 区画ID
//
// 戻り値:
//   - *model.PlotAssignment: アクティブな配置（UnassignedDateがNULL）
//   - error: アクティブな配置がない場合は gorm.ErrRecordNotFound
func (s *Service) GetActivePlotAssignment(ctx context.Context, plotID uint) (*model.PlotAssignment, error) {
	return s.repos.PlotAssignment().GetActiveByPlotID(ctx, plotID)
}

// GetCropAssignments は作物の全配置履歴を取得します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - cropID: 作物ID
//
// 戻り値:
//   - []model.PlotAssignment: 配置履歴の一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetCropAssignments(ctx context.Context, cropID uint) ([]model.PlotAssignment, error) {
	return s.repos.PlotAssignment().GetByCropID(ctx, cropID)
}

// =============================================================================
// Plot Layout & History Methods - 区画レイアウト・履歴メソッド
// =============================================================================

// PlotLayoutItem はレイアウト表示用の区画データです。
// 区画情報と現在の配置情報を含みます。
type PlotLayoutItem struct {
	Plot             model.Plot            `json:"plot"`
	ActiveAssignment *model.PlotAssignment `json:"active_assignment,omitempty"`
	ActiveCrop       *model.Crop           `json:"active_crop,omitempty"`
}

// GetPlotLayout はユーザーの全区画のレイアウトデータを取得します。
// グリッド表示用に、区画情報と現在の配置情報を含むデータを返します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//
// 戻り値:
//   - []PlotLayoutItem: レイアウトデータの一覧
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetPlotLayout(ctx context.Context, userID uint) ([]PlotLayoutItem, error) {
	// 全区画を取得
	plots, err := s.repos.Plot().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// レイアウトデータを構築
	layoutItems := make([]PlotLayoutItem, len(plots))
	for i, plot := range plots {
		item := PlotLayoutItem{
			Plot: plot,
		}

		// アクティブな配置を取得（エラーは無視 - 配置がない場合も正常）
		assignment, err := s.repos.PlotAssignment().GetActiveByPlotID(ctx, plot.ID)
		if err == nil && assignment != nil {
			item.ActiveAssignment = assignment

			// 配置されている作物を取得
			crop, err := s.repos.Crop().GetByID(ctx, assignment.CropID)
			if err == nil {
				item.ActiveCrop = crop
			}
		}

		layoutItems[i] = item
	}

	return layoutItems, nil
}

// PlotHistoryItem は区画履歴表示用のデータです。
// 配置情報と作物情報を含みます。
type PlotHistoryItem struct {
	Assignment model.PlotAssignment `json:"assignment"`
	Crop       *model.Crop          `json:"crop,omitempty"`
}

// GetPlotHistory は区画の栽培履歴を取得します。
// 過去に配置された作物の履歴を返します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - plotID: 区画ID
//
// 戻り値:
//   - []PlotHistoryItem: 履歴データの一覧（配置日の降順）
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetPlotHistory(ctx context.Context, plotID uint) ([]PlotHistoryItem, error) {
	// 全配置履歴を取得
	assignments, err := s.repos.PlotAssignment().GetByPlotID(ctx, plotID)
	if err != nil {
		return nil, err
	}

	// 履歴データを構築
	historyItems := make([]PlotHistoryItem, len(assignments))
	for i, assignment := range assignments {
		item := PlotHistoryItem{
			Assignment: assignment,
		}

		// 作物情報を取得
		crop, err := s.repos.Crop().GetByID(ctx, assignment.CropID)
		if err == nil {
			item.Crop = crop
		}

		historyItems[i] = item
	}

	return historyItems, nil
}

// =============================================================================
// Analytics Service Methods - 分析サービスメソッド
// =============================================================================
// 収穫量の集計やグラフデータの生成を行います。

// HarvestSummary は収穫量集計の結果を表します。
type HarvestSummary struct {
	TotalHarvests      int                `json:"total_harvests"`       // 総収穫回数
	TotalQuantityKg    float64            `json:"total_quantity_kg"`    // 総収穫量（kg換算）
	CropSummaries      []CropHarvestSummary `json:"crop_summaries"`     // 作物ごとの集計
	QualityDistribution map[string]int    `json:"quality_distribution"` // 品質別の分布
}

// CropHarvestSummary は作物ごとの収穫集計を表します。
type CropHarvestSummary struct {
	CropID            uint    `json:"crop_id"`
	CropName          string  `json:"crop_name"`
	HarvestCount      int     `json:"harvest_count"`       // 収穫回数
	TotalQuantity     float64 `json:"total_quantity"`      // 総収穫量
	QuantityUnit      string  `json:"quantity_unit"`       // 数量単位
	TotalQuantityKg   float64 `json:"total_quantity_kg"`   // kg換算の総収穫量
	AverageQuantity   float64 `json:"average_quantity"`    // 平均収穫量
	AverageGrowthDays int     `json:"average_growth_days"` // 平均成長日数
}

// HarvestFilter は収穫データのフィルタ条件を表します。
type HarvestFilter struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CropID    *uint      `json:"crop_id,omitempty"`
}

// GetHarvestSummary はユーザーの収穫量集計を取得します。
// フィルタ条件に基づいて、作物ごとの総収穫量・平均成長期間を集計します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//   - filter: フィルタ条件（日付範囲、作物ID）
//
// 戻り値:
//   - *HarvestSummary: 集計結果
//   - error: 取得に失敗した場合のエラー
func (s *Service) GetHarvestSummary(ctx context.Context, userID uint, filter HarvestFilter) (*HarvestSummary, error) {
	// 収穫データを取得
	harvests, err := s.repos.Harvest().GetByUserIDWithDateRange(ctx, userID, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, err
	}

	// 作物情報を取得するためのマップ
	cropCache := make(map[uint]*model.Crop)

	// 作物IDでフィルタ
	if filter.CropID != nil {
		var filtered []model.Harvest
		for _, h := range harvests {
			if h.CropID == *filter.CropID {
				filtered = append(filtered, h)
			}
		}
		harvests = filtered
	}

	// 作物ごとに集計
	cropStats := make(map[uint]*CropHarvestSummary)
	qualityDist := make(map[string]int)

	for _, harvest := range harvests {
		// 作物情報をキャッシュから取得または取得
		crop, ok := cropCache[harvest.CropID]
		if !ok {
			crop, err = s.repos.Crop().GetByID(ctx, harvest.CropID)
			if err != nil {
				continue // 作物が見つからない場合はスキップ
			}
			cropCache[harvest.CropID] = crop
		}

		// 作物ごとの集計を更新
		stats, ok := cropStats[harvest.CropID]
		if !ok {
			stats = &CropHarvestSummary{
				CropID:       harvest.CropID,
				CropName:     crop.Name,
				QuantityUnit: harvest.QuantityUnit,
			}
			cropStats[harvest.CropID] = stats
		}

		stats.HarvestCount++
		stats.TotalQuantity += harvest.Quantity
		stats.TotalQuantityKg += convertToKg(harvest.Quantity, harvest.QuantityUnit)

		// 成長日数を計算（植え付け日から収穫日まで）
		if !crop.PlantedDate.IsZero() {
			growthDays := int(harvest.HarvestDate.Sub(crop.PlantedDate).Hours() / 24)
			if growthDays > 0 {
				stats.AverageGrowthDays = (stats.AverageGrowthDays*(stats.HarvestCount-1) + growthDays) / stats.HarvestCount
			}
		}

		// 品質分布を更新
		if harvest.Quality != "" {
			qualityDist[harvest.Quality]++
		}
	}

	// 平均収穫量を計算
	var cropSummaries []CropHarvestSummary
	var totalKg float64
	for _, stats := range cropStats {
		if stats.HarvestCount > 0 {
			stats.AverageQuantity = stats.TotalQuantity / float64(stats.HarvestCount)
		}
		cropSummaries = append(cropSummaries, *stats)
		totalKg += stats.TotalQuantityKg
	}

	return &HarvestSummary{
		TotalHarvests:       len(harvests),
		TotalQuantityKg:     totalKg,
		CropSummaries:       cropSummaries,
		QualityDistribution: qualityDist,
	}, nil
}

// convertToKg は指定された単位の数量をkg単位に換算します。
// pieces（個数）の場合は、1個=0.1kgとして概算します。
func convertToKg(quantity float64, unit string) float64 {
	switch unit {
	case "kg":
		return quantity
	case "g":
		return quantity / 1000
	case "pieces":
		// 1個=0.1kg（100g）として概算
		return quantity * 0.1
	default:
		return quantity
	}
}
