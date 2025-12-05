package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	Users         map[uint]*model.User
	UsersByEmail  map[string]*model.User
	NextID        uint
	CreateFunc    func(ctx context.Context, user *model.User) error
	GetByIDFunc   func(ctx context.Context, id uint) (*model.User, error)
	GetByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	UpdateFunc    func(ctx context.Context, user *model.User) error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:        make(map[uint]*model.User),
		UsersByEmail: make(map[string]*model.User),
		NextID:       1,
	}
}

func (r *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, user)
	}
	user.ID = r.NextID
	r.NextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.Users[user.ID] = user
	r.UsersByEmail[user.Email] = user
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	if r.GetByIDFunc != nil {
		return r.GetByIDFunc(ctx, id)
	}
	if user, ok := r.Users[id]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockUserRepository) GetByFirebaseUID(ctx context.Context, uid string) (*model.User, error) {
	for _, user := range r.Users {
		if user.FirebaseUID == uid {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if r.GetByEmailFunc != nil {
		return r.GetByEmailFunc(ctx, email)
	}
	if user, ok := r.UsersByEmail[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, user)
	}
	user.UpdatedAt = time.Now()
	r.Users[user.ID] = user
	r.UsersByEmail[user.Email] = user
	return nil
}

func (r *MockUserRepository) Delete(ctx context.Context, id uint) error {
	if user, ok := r.Users[id]; ok {
		delete(r.UsersByEmail, user.Email)
		delete(r.Users, id)
	}
	return nil
}

// MockTokenBlacklistRepository is a mock implementation of TokenBlacklistRepository
type MockTokenBlacklistRepository struct {
	Tokens map[string]time.Time
}

func NewMockTokenBlacklistRepository() *MockTokenBlacklistRepository {
	return &MockTokenBlacklistRepository{
		Tokens: make(map[string]time.Time),
	}
}

func (r *MockTokenBlacklistRepository) Add(ctx context.Context, tokenHash string, expiresAt time.Time) error {
	r.Tokens[tokenHash] = expiresAt
	return nil
}

func (r *MockTokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	_, exists := r.Tokens[tokenHash]
	return exists, nil
}

func (r *MockTokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for hash, expiresAt := range r.Tokens {
		if expiresAt.Before(now) {
			delete(r.Tokens, hash)
		}
	}
	return nil
}

// MockGardenRepository is a mock implementation of GardenRepository
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

// MockPlantRepository is a mock implementation of PlantRepository
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

// MockCareLogRepository is a mock implementation of CareLogRepository
type MockCareLogRepository struct{}

func (r *MockCareLogRepository) Create(ctx context.Context, careLog *model.CareLog) error { return nil }
func (r *MockCareLogRepository) GetByID(ctx context.Context, id uint) (*model.CareLog, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *MockCareLogRepository) GetByPlantID(ctx context.Context, plantID uint) ([]model.CareLog, error) {
	return nil, nil
}
func (r *MockCareLogRepository) Delete(ctx context.Context, id uint) error { return nil }

// MockRepositories implements Repositories interface for testing
type MockRepositories struct {
	userRepo           *MockUserRepository
	gardenRepo         *MockGardenRepository
	plantRepo          *MockPlantRepository
	careLogRepo        *MockCareLogRepository
	tokenBlacklistRepo *MockTokenBlacklistRepository
}

func NewMockRepositories() *MockRepositories {
	return &MockRepositories{
		userRepo:           NewMockUserRepository(),
		gardenRepo:         &MockGardenRepository{},
		plantRepo:          &MockPlantRepository{},
		careLogRepo:        &MockCareLogRepository{},
		tokenBlacklistRepo: NewMockTokenBlacklistRepository(),
	}
}

func (m *MockRepositories) User() UserRepository {
	return m.userRepo
}

func (m *MockRepositories) Garden() GardenRepository {
	return m.gardenRepo
}

func (m *MockRepositories) Plant() PlantRepository {
	return m.plantRepo
}

func (m *MockRepositories) CareLog() CareLogRepository {
	return m.careLogRepo
}

func (m *MockRepositories) TokenBlacklist() TokenBlacklistRepository {
	return m.tokenBlacklistRepo
}

func (m *MockRepositories) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// GetMockUserRepository returns the underlying mock user repository for test setup
func (m *MockRepositories) GetMockUserRepository() *MockUserRepository {
	return m.userRepo
}

// GetMockTokenBlacklistRepository returns the underlying mock token blacklist repository
func (m *MockRepositories) GetMockTokenBlacklistRepository() *MockTokenBlacklistRepository {
	return m.tokenBlacklistRepo
}
