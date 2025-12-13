// Package repository - Mock Repository
//
// このファイルはユニットテスト用のモックリポジトリを提供します。
// 実際のデータベース（PostgreSQL）の代わりにメモリ内のMapを使用することで、
// 高速で独立したテストを実現します。
//
// 使用例:
//
//	mockRepos := repository.NewMockRepositories()
//	svc := service.NewService(mockRepos)
//	// テストでサービス層を使用
//
// カスタム動作の例（エラーを強制発生させたい場合）:
//
//	mockRepos.GetMockUserRepository().CreateFunc = func(ctx context.Context, user *model.User) error {
//	    return errors.New("database error")
//	}
package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// MockUserRepository は UserRepository インターフェースのモック実装です。
// テスト時にデータベースの代わりにメモリ内のMapを使用します。
type MockUserRepository struct {
	// Users はIDをキーとしたユーザーの格納Map
	// PostgreSQLのプライマリキー検索をシミュレート
	Users map[uint]*model.User

	// UsersByEmail はEmailをキーとしたユーザーの格納Map
	// PostgreSQLのユニークインデックス検索をシミュレート
	// 同じユーザーを2つのMapに格納することで、両方の検索パターンに対応
	UsersByEmail map[string]*model.User

	// NextID は次に割り当てるID（自動インクリメントをシミュレート）
	// PostgreSQLの SERIAL / BIGSERIAL と同じ動作
	NextID uint

	// 以下のFunc系フィールドは、テストでカスタム動作を注入するためのフック
	// nilの場合はデフォルト動作、関数をセットすると優先実行される

	// CreateFunc - Create時のカスタム動作（エラー発生テスト等に使用）
	CreateFunc func(ctx context.Context, user *model.User) error

	// GetByIDFunc - GetByID時のカスタム動作
	GetByIDFunc func(ctx context.Context, id uint) (*model.User, error)

	// GetByEmailFunc - GetByEmail時のカスタム動作
	GetByEmailFunc func(ctx context.Context, email string) (*model.User, error)

	// UpdateFunc - Update時のカスタム動作
	UpdateFunc func(ctx context.Context, user *model.User) error
}

// NewMockUserRepository は新しいMockUserRepositoryを作成します。
// 空のMapを初期化し、IDカウンターを1から開始します。
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:        make(map[uint]*model.User),
		UsersByEmail: make(map[string]*model.User),
		NextID:       1,
	}
}

// Create は新しいユーザーをメモリに保存します。
// PostgreSQLのINSERT文をシミュレートします。
//
// 動作:
//  1. CreateFuncがセットされていれば優先実行（カスタム動作用）
//  2. 自動インクリメントIDを割り当て
//  3. CreatedAt/UpdatedAtを現在時刻に設定（GORMの自動設定をシミュレート）
//  4. 両方のMapに保存（ID検索とEmail検索の両方に対応）
func (r *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	// カスタム関数があれば優先実行（テストでエラーを強制発生させたい場合等）
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, user)
	}

	// 自動インクリメントIDをシミュレート
	user.ID = r.NextID
	r.NextID++

	// GORMのCreatedAt/UpdatedAt自動設定をシミュレート
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 両方のMapに保存（同じポインタを格納）
	r.Users[user.ID] = user
	r.UsersByEmail[user.Email] = user

	return nil
}

// GetByID はIDでユーザーを検索します。
// PostgreSQLの「SELECT * FROM users WHERE id = ?」をシミュレートします。
func (r *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	// カスタム関数があれば優先実行
	if r.GetByIDFunc != nil {
		return r.GetByIDFunc(ctx, id)
	}

	// Mapから検索（O(1)の計算量）
	if user, ok := r.Users[id]; ok {
		return user, nil
	}

	// GORMと同じエラーを返す（一貫性のため）
	return nil, gorm.ErrRecordNotFound
}

// GetByFirebaseUID はFirebase UIDでユーザーを検索します。
// FirebaseUID用のMapがないため、線形探索（O(n)）で検索します。
// テストデータは少量なので、パフォーマンス上問題ありません。
func (r *MockUserRepository) GetByFirebaseUID(ctx context.Context, uid string) (*model.User, error) {
	// 全ユーザーをスキャンして検索
	for _, user := range r.Users {
		if user.FirebaseUID == uid {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByEmail はEmailでユーザーを検索します。
// PostgreSQLの「SELECT * FROM users WHERE email = ?」をシミュレートします。
// UsersByEmailマップを使用してO(1)で検索します。
func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// カスタム関数があれば優先実行
	if r.GetByEmailFunc != nil {
		return r.GetByEmailFunc(ctx, email)
	}

	// Emailインデックス用Mapから検索
	if user, ok := r.UsersByEmail[email]; ok {
		return user, nil
	}

	return nil, gorm.ErrRecordNotFound
}

// Update はユーザー情報を更新します。
// PostgreSQLのUPDATE文をシミュレートします。
//
// 注意: userはポインタなので、呼び出し元の変数も更新されます。
// Mapに再格納しているのは、コードの意図を明確にするためです。
func (r *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	// カスタム関数があれば優先実行
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, user)
	}

	// GORMのUpdatedAt自動更新をシミュレート
	user.UpdatedAt = time.Now()

	// Mapを更新（ポインタなので実際は同じオブジェクト）
	r.Users[user.ID] = user
	r.UsersByEmail[user.Email] = user

	return nil
}

// Delete はユーザーを削除します。
// 両方のMapから削除します（物理削除をシミュレート）。
func (r *MockUserRepository) Delete(ctx context.Context, id uint) error {
	if user, ok := r.Users[id]; ok {
		// 両方のMapから削除
		delete(r.UsersByEmail, user.Email)
		delete(r.Users, id)
	}
	return nil
}

// MockTokenBlacklistRepository は TokenBlacklistRepository のモック実装です。
// ログアウト時のトークン無効化機能をテストするために使用します。
type MockTokenBlacklistRepository struct {
	// Tokens はトークンハッシュをキー、有効期限を値とするMap
	// 「このトークンは無効化されている」という状態を保持
	Tokens map[string]time.Time
}

// NewMockTokenBlacklistRepository は新しいモックを作成します。
func NewMockTokenBlacklistRepository() *MockTokenBlacklistRepository {
	return &MockTokenBlacklistRepository{
		Tokens: make(map[string]time.Time),
	}
}

// Add はトークンをブラックリストに追加します。
// ログアウト時に呼び出され、そのトークンを無効化します。
func (r *MockTokenBlacklistRepository) Add(ctx context.Context, tokenHash string, expiresAt time.Time) error {
	r.Tokens[tokenHash] = expiresAt
	return nil
}

// IsBlacklisted はトークンがブラックリストに登録されているか確認します。
// JWT検証時に呼び出され、ログアウト済みトークンを拒否するために使用します。
func (r *MockTokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	_, exists := r.Tokens[tokenHash]
	return exists, nil
}

// DeleteExpired は期限切れのトークンを削除します。
// 定期的なクリーンアップジョブをシミュレートします。
func (r *MockTokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for hash, expiresAt := range r.Tokens {
		if expiresAt.Before(now) {
			delete(r.Tokens, hash)
		}
	}
	return nil
}

// MockGardenRepository は GardenRepository のスタブ実装です。
type MockGardenRepository struct{}

func (r *MockGardenRepository) Create(ctx context.Context, garden *model.Garden) error { return nil }
func (r *MockGardenRepository) GetByID(ctx context.Context, id uint) (*model.Garden, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *MockGardenRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Garden, error) {
	return nil, nil
}
func (r *MockGardenRepository) Update(ctx context.Context, garden *model.Garden) error { return nil }
func (r *MockGardenRepository) Delete(ctx context.Context, id uint) error              { return nil }

// MockPlantRepository は PlantRepository のスタブ実装です。
type MockPlantRepository struct{}

func (r *MockPlantRepository) Create(ctx context.Context, plant *model.Plant) error { return nil }
func (r *MockPlantRepository) GetByID(ctx context.Context, id uint) (*model.Plant, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *MockPlantRepository) GetByGardenID(ctx context.Context, gardenID uint) ([]model.Plant, error) {
	return nil, nil
}
func (r *MockPlantRepository) Update(ctx context.Context, plant *model.Plant) error { return nil }
func (r *MockPlantRepository) Delete(ctx context.Context, id uint) error             { return nil }
func (r *MockPlantRepository) DeleteByGardenID(ctx context.Context, gardenID uint) error {
	return nil
}

// MockCareLogRepository は CareLogRepository のスタブ実装です。
type MockCareLogRepository struct{}

func (r *MockCareLogRepository) Create(ctx context.Context, careLog *model.CareLog) error { return nil }
func (r *MockCareLogRepository) GetByID(ctx context.Context, id uint) (*model.CareLog, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *MockCareLogRepository) GetByPlantID(ctx context.Context, plantID uint) ([]model.CareLog, error) {
	return nil, nil
}
func (r *MockCareLogRepository) Delete(ctx context.Context, id uint) error { return nil }

// MockTaskRepository は TaskRepository インターフェースのモック実装です。
// タスク管理機能のテストに使用します。
type MockTaskRepository struct {
	// Tasks はIDをキーとしたタスクの格納Map
	Tasks map[uint]*model.Task

	// TasksByUserID はユーザーIDをキーとしたタスクリストの格納Map
	// ユーザーごとのタスク一覧取得をO(1)で実現
	TasksByUserID map[uint][]*model.Task

	// NextID は次に割り当てるID（自動インクリメントをシミュレート）
	NextID uint

	// カスタム動作用のフック関数
	CreateFunc             func(ctx context.Context, task *model.Task) error
	GetByIDFunc            func(ctx context.Context, id uint) (*model.Task, error)
	GetByUserIDFunc        func(ctx context.Context, userID uint) ([]model.Task, error)
	GetByUserIDAndStatusFunc func(ctx context.Context, userID uint, status string) ([]model.Task, error)
	UpdateFunc             func(ctx context.Context, task *model.Task) error
	DeleteFunc             func(ctx context.Context, id uint) error
}

// NewMockTaskRepository は新しいMockTaskRepositoryを作成します。
func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		Tasks:         make(map[uint]*model.Task),
		TasksByUserID: make(map[uint][]*model.Task),
		NextID:        1,
	}
}

// Create は新しいタスクをメモリに保存します。
func (r *MockTaskRepository) Create(ctx context.Context, task *model.Task) error {
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, task)
	}

	task.ID = r.NextID
	r.NextID++
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	r.Tasks[task.ID] = task
	r.TasksByUserID[task.UserID] = append(r.TasksByUserID[task.UserID], task)

	return nil
}

// GetByID はIDでタスクを検索します。
func (r *MockTaskRepository) GetByID(ctx context.Context, id uint) (*model.Task, error) {
	if r.GetByIDFunc != nil {
		return r.GetByIDFunc(ctx, id)
	}

	if task, ok := r.Tasks[id]; ok {
		return task, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByUserID はユーザーIDで全タスクを取得します。
func (r *MockTaskRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Task, error) {
	if r.GetByUserIDFunc != nil {
		return r.GetByUserIDFunc(ctx, userID)
	}

	tasks := r.TasksByUserID[userID]
	result := make([]model.Task, len(tasks))
	for i, t := range tasks {
		result[i] = *t
	}
	return result, nil
}

// GetByUserIDAndStatus はユーザーIDとステータスでタスクを取得します。
func (r *MockTaskRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Task, error) {
	if r.GetByUserIDAndStatusFunc != nil {
		return r.GetByUserIDAndStatusFunc(ctx, userID, status)
	}

	var result []model.Task
	for _, t := range r.TasksByUserID[userID] {
		if t.Status == status {
			result = append(result, *t)
		}
	}
	return result, nil
}

// GetTodayTasks は今日が期限のタスクを取得します。
func (r *MockTaskRepository) GetTodayTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var result []model.Task
	for _, t := range r.TasksByUserID[userID] {
		if t.Status == "pending" && !t.DueDate.Before(today) && t.DueDate.Before(tomorrow) {
			result = append(result, *t)
		}
	}
	return result, nil
}

// GetOverdueTasks は期限切れのタスクを取得します。
func (r *MockTaskRepository) GetOverdueTasks(ctx context.Context, userID uint) ([]model.Task, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var result []model.Task
	for _, t := range r.TasksByUserID[userID] {
		if t.Status == "pending" && t.DueDate.Before(today) {
			result = append(result, *t)
		}
	}
	return result, nil
}

// GetAllOverdueTasks はシステム全体の期限切れタスクを取得します（通知処理用）。
func (r *MockTaskRepository) GetAllOverdueTasks(ctx context.Context) ([]model.Task, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var result []model.Task
	for _, t := range r.Tasks {
		if t.Status == "pending" && t.DueDate.Before(today) {
			result = append(result, *t)
		}
	}
	return result, nil
}

// GetAllTodayTasks はシステム全体の今日が期限のタスクを取得します（通知処理用）。
func (r *MockTaskRepository) GetAllTodayTasks(ctx context.Context) ([]model.Task, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var result []model.Task
	for _, t := range r.Tasks {
		if t.Status == "pending" && !t.DueDate.Before(today) && t.DueDate.Before(tomorrow) {
			result = append(result, *t)
		}
	}
	return result, nil
}

// Update はタスクを更新します。
func (r *MockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, task)
	}

	task.UpdatedAt = time.Now()
	r.Tasks[task.ID] = task
	return nil
}

// Delete はタスクを削除します。
func (r *MockTaskRepository) Delete(ctx context.Context, id uint) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(ctx, id)
	}

	if task, ok := r.Tasks[id]; ok {
		// TasksByUserIDからも削除
		userTasks := r.TasksByUserID[task.UserID]
		for i, t := range userTasks {
			if t.ID == id {
				r.TasksByUserID[task.UserID] = append(userTasks[:i], userTasks[i+1:]...)
				break
			}
		}
		delete(r.Tasks, id)
	}
	return nil
}

// MockCropRepository は CropRepository インターフェースのモック実装です。
// 作物管理機能のテストに使用します。
type MockCropRepository struct {
	// Crops はIDをキーとした作物の格納Map
	Crops map[uint]*model.Crop

	// CropsByUserID はユーザーIDをキーとした作物リストの格納Map
	CropsByUserID map[uint][]*model.Crop

	// NextID は次に割り当てるID
	NextID uint

	// カスタム動作用のフック関数
	CreateFunc             func(ctx context.Context, crop *model.Crop) error
	GetByIDFunc            func(ctx context.Context, id uint) (*model.Crop, error)
	GetByUserIDFunc        func(ctx context.Context, userID uint) ([]model.Crop, error)
	GetByUserIDAndStatusFunc func(ctx context.Context, userID uint, status string) ([]model.Crop, error)
	UpdateFunc             func(ctx context.Context, crop *model.Crop) error
	DeleteFunc             func(ctx context.Context, id uint) error
}

// NewMockCropRepository は新しいMockCropRepositoryを作成します。
func NewMockCropRepository() *MockCropRepository {
	return &MockCropRepository{
		Crops:         make(map[uint]*model.Crop),
		CropsByUserID: make(map[uint][]*model.Crop),
		NextID:        1,
	}
}

// Create は新しい作物をメモリに保存します。
func (r *MockCropRepository) Create(ctx context.Context, crop *model.Crop) error {
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, crop)
	}

	crop.ID = r.NextID
	r.NextID++
	crop.CreatedAt = time.Now()
	crop.UpdatedAt = time.Now()

	r.Crops[crop.ID] = crop
	r.CropsByUserID[crop.UserID] = append(r.CropsByUserID[crop.UserID], crop)

	return nil
}

// GetByID はIDで作物を検索します。
func (r *MockCropRepository) GetByID(ctx context.Context, id uint) (*model.Crop, error) {
	if r.GetByIDFunc != nil {
		return r.GetByIDFunc(ctx, id)
	}

	if crop, ok := r.Crops[id]; ok {
		return crop, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByUserID はユーザーIDで全作物を取得します。
func (r *MockCropRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Crop, error) {
	if r.GetByUserIDFunc != nil {
		return r.GetByUserIDFunc(ctx, userID)
	}

	crops := r.CropsByUserID[userID]
	result := make([]model.Crop, len(crops))
	for i, c := range crops {
		result[i] = *c
	}
	return result, nil
}

// GetByUserIDAndStatus はユーザーIDとステータスで作物を取得します。
func (r *MockCropRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Crop, error) {
	if r.GetByUserIDAndStatusFunc != nil {
		return r.GetByUserIDAndStatusFunc(ctx, userID, status)
	}

	var result []model.Crop
	for _, c := range r.CropsByUserID[userID] {
		if c.Status == status {
			result = append(result, *c)
		}
	}
	return result, nil
}

// GetUpcomingHarvests は指定日数以内に収穫予定の作物を取得します（通知処理用）。
func (r *MockCropRepository) GetUpcomingHarvests(ctx context.Context, daysAhead int) ([]model.Crop, error) {
	today := time.Now().Truncate(24 * time.Hour)
	targetDate := today.AddDate(0, 0, daysAhead)

	var result []model.Crop
	for _, c := range r.Crops {
		if c.Status == "growing" &&
			!c.ExpectedHarvestDate.Before(today) &&
			!c.ExpectedHarvestDate.After(targetDate) {
			result = append(result, *c)
		}
	}
	return result, nil
}

// Update は作物を更新します。
func (r *MockCropRepository) Update(ctx context.Context, crop *model.Crop) error {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, crop)
	}

	crop.UpdatedAt = time.Now()
	r.Crops[crop.ID] = crop
	return nil
}

// Delete は作物を削除します。
func (r *MockCropRepository) Delete(ctx context.Context, id uint) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(ctx, id)
	}

	if crop, ok := r.Crops[id]; ok {
		// CropsByUserIDからも削除
		userCrops := r.CropsByUserID[crop.UserID]
		for i, c := range userCrops {
			if c.ID == id {
				r.CropsByUserID[crop.UserID] = append(userCrops[:i], userCrops[i+1:]...)
				break
			}
		}
		delete(r.Crops, id)
	}
	return nil
}

// MockGrowthRecordRepository は GrowthRecordRepository インターフェースのモック実装です。
type MockGrowthRecordRepository struct {
	// Records はIDをキーとした成長記録の格納Map
	Records map[uint]*model.GrowthRecord

	// RecordsByCropID は作物IDをキーとした成長記録リストの格納Map
	RecordsByCropID map[uint][]*model.GrowthRecord

	// NextID は次に割り当てるID
	NextID uint
}

// NewMockGrowthRecordRepository は新しいMockGrowthRecordRepositoryを作成します。
func NewMockGrowthRecordRepository() *MockGrowthRecordRepository {
	return &MockGrowthRecordRepository{
		Records:         make(map[uint]*model.GrowthRecord),
		RecordsByCropID: make(map[uint][]*model.GrowthRecord),
		NextID:          1,
	}
}

// Create は新しい成長記録をメモリに保存します。
func (r *MockGrowthRecordRepository) Create(ctx context.Context, record *model.GrowthRecord) error {
	record.ID = r.NextID
	r.NextID++
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	r.Records[record.ID] = record
	r.RecordsByCropID[record.CropID] = append(r.RecordsByCropID[record.CropID], record)

	return nil
}

// GetByID はIDで成長記録を検索します。
func (r *MockGrowthRecordRepository) GetByID(ctx context.Context, id uint) (*model.GrowthRecord, error) {
	if record, ok := r.Records[id]; ok {
		return record, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByCropID は作物IDで全成長記録を取得します。
func (r *MockGrowthRecordRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.GrowthRecord, error) {
	records := r.RecordsByCropID[cropID]
	result := make([]model.GrowthRecord, len(records))
	for i, rec := range records {
		result[i] = *rec
	}
	return result, nil
}

// Delete は成長記録を削除します。
func (r *MockGrowthRecordRepository) Delete(ctx context.Context, id uint) error {
	if record, ok := r.Records[id]; ok {
		// RecordsByCropIDからも削除
		cropRecords := r.RecordsByCropID[record.CropID]
		for i, rec := range cropRecords {
			if rec.ID == id {
				r.RecordsByCropID[record.CropID] = append(cropRecords[:i], cropRecords[i+1:]...)
				break
			}
		}
		delete(r.Records, id)
	}
	return nil
}

// DeleteByCropID は作物IDで全成長記録を削除します（バッチ削除）。
func (r *MockGrowthRecordRepository) DeleteByCropID(ctx context.Context, cropID uint) error {
	for _, record := range r.RecordsByCropID[cropID] {
		delete(r.Records, record.ID)
	}
	delete(r.RecordsByCropID, cropID)
	return nil
}

// MockHarvestRepository は HarvestRepository インターフェースのモック実装です。
type MockHarvestRepository struct {
	// Harvests はIDをキーとした収穫記録の格納Map
	Harvests map[uint]*model.Harvest

	// HarvestsByCropID は作物IDをキーとした収穫記録リストの格納Map
	HarvestsByCropID map[uint][]*model.Harvest

	// HarvestsByUserID はユーザーIDをキーとした収穫記録リストの格納Map（Analytics用）
	HarvestsByUserID map[uint][]*model.Harvest

	// NextID は次に割り当てるID
	NextID uint

	// カスタム動作用のフック関数
	GetByUserIDWithDateRangeFunc func(ctx context.Context, userID uint, startDate, endDate *time.Time) ([]model.Harvest, error)
}

// NewMockHarvestRepository は新しいMockHarvestRepositoryを作成します。
func NewMockHarvestRepository() *MockHarvestRepository {
	return &MockHarvestRepository{
		Harvests:         make(map[uint]*model.Harvest),
		HarvestsByCropID: make(map[uint][]*model.Harvest),
		HarvestsByUserID: make(map[uint][]*model.Harvest),
		NextID:           1,
	}
}

// Create は新しい収穫記録をメモリに保存します。
func (r *MockHarvestRepository) Create(ctx context.Context, harvest *model.Harvest) error {
	harvest.ID = r.NextID
	r.NextID++
	harvest.CreatedAt = time.Now()
	harvest.UpdatedAt = time.Now()

	r.Harvests[harvest.ID] = harvest
	r.HarvestsByCropID[harvest.CropID] = append(r.HarvestsByCropID[harvest.CropID], harvest)

	return nil
}

// GetByID はIDで収穫記録を検索します。
func (r *MockHarvestRepository) GetByID(ctx context.Context, id uint) (*model.Harvest, error) {
	if harvest, ok := r.Harvests[id]; ok {
		return harvest, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByCropID は作物IDで全収穫記録を取得します。
func (r *MockHarvestRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.Harvest, error) {
	harvests := r.HarvestsByCropID[cropID]
	result := make([]model.Harvest, len(harvests))
	for i, h := range harvests {
		result[i] = *h
	}
	return result, nil
}

// Delete は収穫記録を削除します。
func (r *MockHarvestRepository) Delete(ctx context.Context, id uint) error {
	if harvest, ok := r.Harvests[id]; ok {
		// HarvestsByCropIDからも削除
		cropHarvests := r.HarvestsByCropID[harvest.CropID]
		for i, h := range cropHarvests {
			if h.ID == id {
				r.HarvestsByCropID[harvest.CropID] = append(cropHarvests[:i], cropHarvests[i+1:]...)
				break
			}
		}
		delete(r.Harvests, id)
	}
	return nil
}

// DeleteByCropID は作物IDで全収穫記録を削除します（バッチ削除）。
func (r *MockHarvestRepository) DeleteByCropID(ctx context.Context, cropID uint) error {
	for _, harvest := range r.HarvestsByCropID[cropID] {
		delete(r.Harvests, harvest.ID)
	}
	delete(r.HarvestsByCropID, cropID)
	return nil
}

// GetByUserIDWithDateRange はユーザーの収穫記録を日付範囲でフィルタして取得します。
// HarvestsByUserIDに事前にデータをセットするか、GetByUserIDWithDateRangeFuncを使用してください。
func (r *MockHarvestRepository) GetByUserIDWithDateRange(ctx context.Context, userID uint, startDate, endDate *time.Time) ([]model.Harvest, error) {
	// カスタム関数が設定されている場合はそれを使用
	if r.GetByUserIDWithDateRangeFunc != nil {
		return r.GetByUserIDWithDateRangeFunc(ctx, userID, startDate, endDate)
	}

	// デフォルト: HarvestsByUserIDからフィルタリング
	harvests := r.HarvestsByUserID[userID]
	var result []model.Harvest
	for _, h := range harvests {
		// 日付範囲フィルタ
		if startDate != nil && h.HarvestDate.Before(*startDate) {
			continue
		}
		if endDate != nil && h.HarvestDate.After(*endDate) {
			continue
		}
		result = append(result, *h)
	}
	return result, nil
}

// AddHarvestForUser はテスト用にユーザーIDに関連付けて収穫記録を追加します。
// Analytics機能のテストで使用します。
func (r *MockHarvestRepository) AddHarvestForUser(userID uint, harvest *model.Harvest) {
	harvest.ID = r.NextID
	r.NextID++
	harvest.CreatedAt = time.Now()
	harvest.UpdatedAt = time.Now()

	r.Harvests[harvest.ID] = harvest
	r.HarvestsByCropID[harvest.CropID] = append(r.HarvestsByCropID[harvest.CropID], harvest)
	r.HarvestsByUserID[userID] = append(r.HarvestsByUserID[userID], harvest)
}

// MockPlotRepository は PlotRepository インターフェースのモック実装です。
// 区画管理機能のテストに使用します。
type MockPlotRepository struct {
	// Plots はIDをキーとした区画の格納Map
	Plots map[uint]*model.Plot

	// PlotsByUserID はユーザーIDをキーとした区画リストの格納Map
	PlotsByUserID map[uint][]*model.Plot

	// NextID は次に割り当てるID
	NextID uint

	// カスタム動作用のフック関数
	CreateFunc               func(ctx context.Context, plot *model.Plot) error
	GetByIDFunc              func(ctx context.Context, id uint) (*model.Plot, error)
	GetByUserIDFunc          func(ctx context.Context, userID uint) ([]model.Plot, error)
	GetByUserIDAndStatusFunc func(ctx context.Context, userID uint, status string) ([]model.Plot, error)
	UpdateFunc               func(ctx context.Context, plot *model.Plot) error
	DeleteFunc               func(ctx context.Context, id uint) error
}

// NewMockPlotRepository は新しいMockPlotRepositoryを作成します。
func NewMockPlotRepository() *MockPlotRepository {
	return &MockPlotRepository{
		Plots:         make(map[uint]*model.Plot),
		PlotsByUserID: make(map[uint][]*model.Plot),
		NextID:        1,
	}
}

// Create は新しい区画をメモリに保存します。
func (r *MockPlotRepository) Create(ctx context.Context, plot *model.Plot) error {
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, plot)
	}

	plot.ID = r.NextID
	r.NextID++
	plot.CreatedAt = time.Now()
	plot.UpdatedAt = time.Now()

	r.Plots[plot.ID] = plot
	r.PlotsByUserID[plot.UserID] = append(r.PlotsByUserID[plot.UserID], plot)

	return nil
}

// GetByID はIDで区画を検索します。
func (r *MockPlotRepository) GetByID(ctx context.Context, id uint) (*model.Plot, error) {
	if r.GetByIDFunc != nil {
		return r.GetByIDFunc(ctx, id)
	}

	if plot, ok := r.Plots[id]; ok {
		return plot, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByUserID はユーザーIDで全区画を取得します。
func (r *MockPlotRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Plot, error) {
	if r.GetByUserIDFunc != nil {
		return r.GetByUserIDFunc(ctx, userID)
	}

	plots := r.PlotsByUserID[userID]
	result := make([]model.Plot, len(plots))
	for i, p := range plots {
		result[i] = *p
	}
	return result, nil
}

// GetByUserIDAndStatus はユーザーIDとステータスで区画を取得します。
func (r *MockPlotRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Plot, error) {
	if r.GetByUserIDAndStatusFunc != nil {
		return r.GetByUserIDAndStatusFunc(ctx, userID, status)
	}

	var result []model.Plot
	for _, p := range r.PlotsByUserID[userID] {
		if p.Status == status {
			result = append(result, *p)
		}
	}
	return result, nil
}

// Update は区画を更新します。
func (r *MockPlotRepository) Update(ctx context.Context, plot *model.Plot) error {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, plot)
	}

	plot.UpdatedAt = time.Now()
	r.Plots[plot.ID] = plot
	return nil
}

// Delete は区画を削除します。
func (r *MockPlotRepository) Delete(ctx context.Context, id uint) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(ctx, id)
	}

	if plot, ok := r.Plots[id]; ok {
		// PlotsByUserIDからも削除
		userPlots := r.PlotsByUserID[plot.UserID]
		for i, p := range userPlots {
			if p.ID == id {
				r.PlotsByUserID[plot.UserID] = append(userPlots[:i], userPlots[i+1:]...)
				break
			}
		}
		delete(r.Plots, id)
	}
	return nil
}

// MockPlotAssignmentRepository は PlotAssignmentRepository インターフェースのモック実装です。
// 区画への作物配置管理機能のテストに使用します。
type MockPlotAssignmentRepository struct {
	// Assignments はIDをキーとした配置の格納Map
	Assignments map[uint]*model.PlotAssignment

	// AssignmentsByPlotID は区画IDをキーとした配置リストの格納Map
	AssignmentsByPlotID map[uint][]*model.PlotAssignment

	// AssignmentsByCropID は作物IDをキーとした配置リストの格納Map
	AssignmentsByCropID map[uint][]*model.PlotAssignment

	// NextID は次に割り当てるID
	NextID uint
}

// NewMockPlotAssignmentRepository は新しいMockPlotAssignmentRepositoryを作成します。
func NewMockPlotAssignmentRepository() *MockPlotAssignmentRepository {
	return &MockPlotAssignmentRepository{
		Assignments:         make(map[uint]*model.PlotAssignment),
		AssignmentsByPlotID: make(map[uint][]*model.PlotAssignment),
		AssignmentsByCropID: make(map[uint][]*model.PlotAssignment),
		NextID:              1,
	}
}

// Create は新しい区画配置をメモリに保存します。
func (r *MockPlotAssignmentRepository) Create(ctx context.Context, assignment *model.PlotAssignment) error {
	assignment.ID = r.NextID
	r.NextID++
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	r.Assignments[assignment.ID] = assignment
	r.AssignmentsByPlotID[assignment.PlotID] = append(r.AssignmentsByPlotID[assignment.PlotID], assignment)
	r.AssignmentsByCropID[assignment.CropID] = append(r.AssignmentsByCropID[assignment.CropID], assignment)

	return nil
}

// GetByID はIDで区画配置を検索します。
func (r *MockPlotAssignmentRepository) GetByID(ctx context.Context, id uint) (*model.PlotAssignment, error) {
	if assignment, ok := r.Assignments[id]; ok {
		return assignment, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByPlotID は区画IDで全配置履歴を取得します。
func (r *MockPlotAssignmentRepository) GetByPlotID(ctx context.Context, plotID uint) ([]model.PlotAssignment, error) {
	assignments := r.AssignmentsByPlotID[plotID]
	result := make([]model.PlotAssignment, len(assignments))
	for i, a := range assignments {
		result[i] = *a
	}
	return result, nil
}

// GetActiveByPlotID は区画の現在アクティブな配置を取得します。
func (r *MockPlotAssignmentRepository) GetActiveByPlotID(ctx context.Context, plotID uint) (*model.PlotAssignment, error) {
	for _, a := range r.AssignmentsByPlotID[plotID] {
		if a.UnassignedDate == nil {
			return a, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByCropID は作物IDで全配置履歴を取得します。
func (r *MockPlotAssignmentRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.PlotAssignment, error) {
	assignments := r.AssignmentsByCropID[cropID]
	result := make([]model.PlotAssignment, len(assignments))
	for i, a := range assignments {
		result[i] = *a
	}
	return result, nil
}

// Update は区画配置を更新します。
func (r *MockPlotAssignmentRepository) Update(ctx context.Context, assignment *model.PlotAssignment) error {
	assignment.UpdatedAt = time.Now()
	r.Assignments[assignment.ID] = assignment
	return nil
}

// Delete は区画配置を削除します。
func (r *MockPlotAssignmentRepository) Delete(ctx context.Context, id uint) error {
	if assignment, ok := r.Assignments[id]; ok {
		// AssignmentsByPlotIDからも削除
		plotAssignments := r.AssignmentsByPlotID[assignment.PlotID]
		for i, a := range plotAssignments {
			if a.ID == id {
				r.AssignmentsByPlotID[assignment.PlotID] = append(plotAssignments[:i], plotAssignments[i+1:]...)
				break
			}
		}
		// AssignmentsByCropIDからも削除
		cropAssignments := r.AssignmentsByCropID[assignment.CropID]
		for i, a := range cropAssignments {
			if a.ID == id {
				r.AssignmentsByCropID[assignment.CropID] = append(cropAssignments[:i], cropAssignments[i+1:]...)
				break
			}
		}
		delete(r.Assignments, id)
	}
	return nil
}

// DeleteByPlotID は区画IDで全配置を削除します（バッチ削除）。
func (r *MockPlotAssignmentRepository) DeleteByPlotID(ctx context.Context, plotID uint) error {
	for _, assignment := range r.AssignmentsByPlotID[plotID] {
		// AssignmentsByCropIDからも削除
		cropAssignments := r.AssignmentsByCropID[assignment.CropID]
		for i, a := range cropAssignments {
			if a.ID == assignment.ID {
				r.AssignmentsByCropID[assignment.CropID] = append(cropAssignments[:i], cropAssignments[i+1:]...)
				break
			}
		}
		delete(r.Assignments, assignment.ID)
	}
	delete(r.AssignmentsByPlotID, plotID)
	return nil
}

// MockDeviceTokenRepository は DeviceTokenRepository インターフェースのモック実装です。
type MockDeviceTokenRepository struct {
	Tokens          map[uint]*model.DeviceToken
	TokensByUserID  map[uint][]*model.DeviceToken
	TokensByToken   map[string]*model.DeviceToken
	NextID          uint
}

// NewMockDeviceTokenRepository は新しいMockDeviceTokenRepositoryを作成します。
func NewMockDeviceTokenRepository() *MockDeviceTokenRepository {
	return &MockDeviceTokenRepository{
		Tokens:         make(map[uint]*model.DeviceToken),
		TokensByUserID: make(map[uint][]*model.DeviceToken),
		TokensByToken:  make(map[string]*model.DeviceToken),
		NextID:         1,
	}
}

func (r *MockDeviceTokenRepository) Create(ctx context.Context, token *model.DeviceToken) error {
	token.ID = r.NextID
	r.NextID++
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	r.Tokens[token.ID] = token
	r.TokensByUserID[token.UserID] = append(r.TokensByUserID[token.UserID], token)
	r.TokensByToken[token.Token] = token
	return nil
}

func (r *MockDeviceTokenRepository) GetByID(ctx context.Context, id uint) (*model.DeviceToken, error) {
	if token, ok := r.Tokens[id]; ok {
		return token, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockDeviceTokenRepository) GetByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error) {
	tokens := r.TokensByUserID[userID]
	result := make([]model.DeviceToken, len(tokens))
	for i, t := range tokens {
		result[i] = *t
	}
	return result, nil
}

func (r *MockDeviceTokenRepository) GetByUserIDAndPlatform(ctx context.Context, userID uint, platform string) (*model.DeviceToken, error) {
	for _, t := range r.TokensByUserID[userID] {
		if t.Platform == platform {
			return t, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockDeviceTokenRepository) GetByToken(ctx context.Context, token string) (*model.DeviceToken, error) {
	if t, ok := r.TokensByToken[token]; ok {
		return t, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockDeviceTokenRepository) GetActiveByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error) {
	var result []model.DeviceToken
	for _, t := range r.TokensByUserID[userID] {
		if t.IsActive {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *MockDeviceTokenRepository) Update(ctx context.Context, token *model.DeviceToken) error {
	token.UpdatedAt = time.Now()
	r.Tokens[token.ID] = token
	return nil
}

func (r *MockDeviceTokenRepository) Delete(ctx context.Context, id uint) error {
	if token, ok := r.Tokens[id]; ok {
		delete(r.TokensByToken, token.Token)
		tokens := r.TokensByUserID[token.UserID]
		for i, t := range tokens {
			if t.ID == id {
				r.TokensByUserID[token.UserID] = append(tokens[:i], tokens[i+1:]...)
				break
			}
		}
		delete(r.Tokens, id)
	}
	return nil
}

func (r *MockDeviceTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	for _, token := range r.TokensByUserID[userID] {
		delete(r.TokensByToken, token.Token)
		delete(r.Tokens, token.ID)
	}
	delete(r.TokensByUserID, userID)
	return nil
}

func (r *MockDeviceTokenRepository) DeactivateToken(ctx context.Context, id uint) error {
	if token, ok := r.Tokens[id]; ok {
		token.IsActive = false
		token.UpdatedAt = time.Now()
	}
	return nil
}

// MockNotificationLogRepository は NotificationLogRepository インターフェースのモック実装です。
type MockNotificationLogRepository struct {
	Logs                 map[uint]*model.NotificationLog
	LogsByUserID         map[uint][]*model.NotificationLog
	LogsByDeduplication  map[string]*model.NotificationLog
	NextID               uint
}

// NewMockNotificationLogRepository は新しいMockNotificationLogRepositoryを作成します。
func NewMockNotificationLogRepository() *MockNotificationLogRepository {
	return &MockNotificationLogRepository{
		Logs:                make(map[uint]*model.NotificationLog),
		LogsByUserID:        make(map[uint][]*model.NotificationLog),
		LogsByDeduplication: make(map[string]*model.NotificationLog),
		NextID:              1,
	}
}

func (r *MockNotificationLogRepository) Create(ctx context.Context, log *model.NotificationLog) error {
	log.ID = r.NextID
	r.NextID++
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()
	r.Logs[log.ID] = log
	r.LogsByUserID[log.UserID] = append(r.LogsByUserID[log.UserID], log)
	if log.DeduplicationKey != "" {
		r.LogsByDeduplication[log.DeduplicationKey] = log
	}
	return nil
}

func (r *MockNotificationLogRepository) GetByID(ctx context.Context, id uint) (*model.NotificationLog, error) {
	if log, ok := r.Logs[id]; ok {
		return log, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockNotificationLogRepository) GetByDeduplicationKey(ctx context.Context, key string) (*model.NotificationLog, error) {
	if log, ok := r.LogsByDeduplication[key]; ok {
		if log.ExpiresAt.After(time.Now()) {
			return log, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockNotificationLogRepository) GetByUserID(ctx context.Context, userID uint, limit int) ([]model.NotificationLog, error) {
	logs := r.LogsByUserID[userID]
	result := make([]model.NotificationLog, 0, len(logs))
	for i := len(logs) - 1; i >= 0 && (limit <= 0 || len(result) < limit); i-- {
		result = append(result, *logs[i])
	}
	return result, nil
}

func (r *MockNotificationLogRepository) GetPendingNotifications(ctx context.Context, limit int) ([]model.NotificationLog, error) {
	var result []model.NotificationLog
	for _, log := range r.Logs {
		if log.Status == "pending" && log.RetryCount < 3 {
			result = append(result, *log)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *MockNotificationLogRepository) Update(ctx context.Context, log *model.NotificationLog) error {
	log.UpdatedAt = time.Now()
	r.Logs[log.ID] = log
	return nil
}

func (r *MockNotificationLogRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for id, log := range r.Logs {
		if log.ExpiresAt.Before(now) {
			delete(r.LogsByDeduplication, log.DeduplicationKey)
			logs := r.LogsByUserID[log.UserID]
			for i, l := range logs {
				if l.ID == id {
					r.LogsByUserID[log.UserID] = append(logs[:i], logs[i+1:]...)
					break
				}
			}
			delete(r.Logs, id)
		}
	}
	return nil
}

// MockRepositories は Repositories インターフェースのモック実装です。
// 各リポジトリのモックを保持し、テストで依存性注入するために使用します。
//
// 設計パターン: ファサードパターン
// - 複数のモックリポジトリを1つのインターフェースでまとめる
// - Service層は本番/テストを意識せずRepositoriesインターフェースを使用
type MockRepositories struct {
	userRepo            *MockUserRepository
	gardenRepo          *MockGardenRepository
	plantRepo           *MockPlantRepository
	careLogRepo         *MockCareLogRepository
	tokenBlacklistRepo  *MockTokenBlacklistRepository
	taskRepo            *MockTaskRepository
	cropRepo            *MockCropRepository
	growthRecordRepo    *MockGrowthRecordRepository
	harvestRepo         *MockHarvestRepository
	plotRepo            *MockPlotRepository
	plotAssignmentRepo  *MockPlotAssignmentRepository
	deviceTokenRepo     *MockDeviceTokenRepository
	notificationLogRepo *MockNotificationLogRepository
}

// NewMockRepositories は新しいMockRepositoriesを作成します。
// 各モックリポジトリを初期化して返します。
func NewMockRepositories() *MockRepositories {
	return &MockRepositories{
		userRepo:            NewMockUserRepository(),
		gardenRepo:          &MockGardenRepository{},
		plantRepo:           &MockPlantRepository{},
		careLogRepo:         &MockCareLogRepository{},
		tokenBlacklistRepo:  NewMockTokenBlacklistRepository(),
		taskRepo:            NewMockTaskRepository(),
		cropRepo:            NewMockCropRepository(),
		growthRecordRepo:    NewMockGrowthRecordRepository(),
		harvestRepo:         NewMockHarvestRepository(),
		plotRepo:            NewMockPlotRepository(),
		plotAssignmentRepo:  NewMockPlotAssignmentRepository(),
		deviceTokenRepo:     NewMockDeviceTokenRepository(),
		notificationLogRepo: NewMockNotificationLogRepository(),
	}
}

// User は UserRepository インターフェースを返します。
// Service層から呼び出されます。
func (m *MockRepositories) User() UserRepository {
	return m.userRepo
}

// Garden は GardenRepository インターフェースを返します。
func (m *MockRepositories) Garden() GardenRepository {
	return m.gardenRepo
}

// Plant は PlantRepository インターフェースを返します。
func (m *MockRepositories) Plant() PlantRepository {
	return m.plantRepo
}

// CareLog は CareLogRepository インターフェースを返します。
func (m *MockRepositories) CareLog() CareLogRepository {
	return m.careLogRepo
}

// TokenBlacklist は TokenBlacklistRepository インターフェースを返します。
func (m *MockRepositories) TokenBlacklist() TokenBlacklistRepository {
	return m.tokenBlacklistRepo
}

// Task は TaskRepository インターフェースを返します。
func (m *MockRepositories) Task() TaskRepository {
	return m.taskRepo
}

// Crop は CropRepository インターフェースを返します。
func (m *MockRepositories) Crop() CropRepository {
	return m.cropRepo
}

// GrowthRecord は GrowthRecordRepository インターフェースを返します。
func (m *MockRepositories) GrowthRecord() GrowthRecordRepository {
	return m.growthRecordRepo
}

// Harvest は HarvestRepository インターフェースを返します。
func (m *MockRepositories) Harvest() HarvestRepository {
	return m.harvestRepo
}

// Plot は PlotRepository インターフェースを返します。
func (m *MockRepositories) Plot() PlotRepository {
	return m.plotRepo
}

// PlotAssignment は PlotAssignmentRepository インターフェースを返します。
func (m *MockRepositories) PlotAssignment() PlotAssignmentRepository {
	return m.plotAssignmentRepo
}

// DeviceToken は DeviceTokenRepository インターフェースを返します。
func (m *MockRepositories) DeviceToken() DeviceTokenRepository {
	return m.deviceTokenRepo
}

// NotificationLog は NotificationLogRepository インターフェースを返します。
func (m *MockRepositories) NotificationLog() NotificationLogRepository {
	return m.notificationLogRepo
}

// WithTransaction はトランザクション処理をシミュレートします。
//
// 本番との違い:
// - 本番: BEGIN → 関数実行 → COMMIT or ROLLBACK
// - モック: 関数を直接実行（トランザクションなし）
//
// テストでこれで問題ない理由:
// - 各テストは独立したMockRepositoriesを作成
// - テスト間でデータが共有されない
// - ロールバックをテストしたい場合はCreateFunc等でエラーを投げる
func (m *MockRepositories) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 単純に関数を実行するだけ（BEGIN/COMMIT/ROLLBACKなし）
	return fn(ctx)
}

// GetMockUserRepository はテストセットアップ用に内部のモックリポジトリを返します。
//
// なぜ必要か:
// - User() は UserRepository インターフェースを返す
// - インターフェース経由では Users Map や CreateFunc にアクセスできない
// - テストでデータをセットアップしたり、カスタム動作を注入するために必要
//
// 使用例:
//
//	mockRepos := repository.NewMockRepositories()
//	mockRepos.GetMockUserRepository().Users[1] = &model.User{...}  // データセットアップ
//	mockRepos.GetMockUserRepository().CreateFunc = func(...) error { ... }  // カスタム動作
func (m *MockRepositories) GetMockUserRepository() *MockUserRepository {
	return m.userRepo
}

// GetMockTokenBlacklistRepository はテスト用に内部のトークンブラックリストモックを返します。
// トークンがブラックリストに登録されたか確認するテストで使用します。
func (m *MockRepositories) GetMockTokenBlacklistRepository() *MockTokenBlacklistRepository {
	return m.tokenBlacklistRepo
}

// GetMockTaskRepository はテスト用に内部のタスクモックを返します。
// タスクのテストデータセットアップやカスタム動作注入に使用します。
func (m *MockRepositories) GetMockTaskRepository() *MockTaskRepository {
	return m.taskRepo
}

// GetMockCropRepository はテスト用に内部の作物モックを返します。
// 作物のテストデータセットアップやカスタム動作注入に使用します。
func (m *MockRepositories) GetMockCropRepository() *MockCropRepository {
	return m.cropRepo
}

// GetMockGrowthRecordRepository はテスト用に内部の成長記録モックを返します。
func (m *MockRepositories) GetMockGrowthRecordRepository() *MockGrowthRecordRepository {
	return m.growthRecordRepo
}

// GetMockHarvestRepository はテスト用に内部の収穫記録モックを返します。
func (m *MockRepositories) GetMockHarvestRepository() *MockHarvestRepository {
	return m.harvestRepo
}

// GetMockPlotRepository はテスト用に内部の区画モックを返します。
// 区画のテストデータセットアップやカスタム動作注入に使用します。
func (m *MockRepositories) GetMockPlotRepository() *MockPlotRepository {
	return m.plotRepo
}

// GetMockPlotAssignmentRepository はテスト用に内部の区画配置モックを返します。
func (m *MockRepositories) GetMockPlotAssignmentRepository() *MockPlotAssignmentRepository {
	return m.plotAssignmentRepo
}
