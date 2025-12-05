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

// =============================================================================
// MockUserRepository - ユーザーリポジトリのモック実装
// =============================================================================

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

// =============================================================================
// MockTokenBlacklistRepository - トークンブラックリストのモック実装
// =============================================================================

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

// =============================================================================
// スタブ実装 - 認証テストでは使用しないリポジトリ
// =============================================================================
// 以下のモックは Repositories インターフェースを満たすために必要ですが、
// 認証テストでは使用しないため、最小限のスタブ実装となっています。
// Garden/Plant/CareLog のテストを書く際に、必要に応じて拡張してください。

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

// =============================================================================
// MockTaskRepository - タスクリポジトリのモック実装
// =============================================================================

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

// =============================================================================
// MockRepositories - 全モックを束ねるファサード
// =============================================================================

// MockRepositories は Repositories インターフェースのモック実装です。
// 各リポジトリのモックを保持し、テストで依存性注入するために使用します。
//
// 設計パターン: ファサードパターン
// - 複数のモックリポジトリを1つのインターフェースでまとめる
// - Service層は本番/テストを意識せずRepositoriesインターフェースを使用
type MockRepositories struct {
	userRepo           *MockUserRepository
	gardenRepo         *MockGardenRepository
	plantRepo          *MockPlantRepository
	careLogRepo        *MockCareLogRepository
	tokenBlacklistRepo *MockTokenBlacklistRepository
	taskRepo           *MockTaskRepository
}

// NewMockRepositories は新しいMockRepositoriesを作成します。
// 各モックリポジトリを初期化して返します。
func NewMockRepositories() *MockRepositories {
	return &MockRepositories{
		userRepo:           NewMockUserRepository(),
		gardenRepo:         &MockGardenRepository{},
		plantRepo:          &MockPlantRepository{},
		careLogRepo:        &MockCareLogRepository{},
		tokenBlacklistRepo: NewMockTokenBlacklistRepository(),
		taskRepo:           NewMockTaskRepository(),
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

// =============================================================================
// テスト用ヘルパーメソッド
// =============================================================================

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
