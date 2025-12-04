package service

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
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
		// Get all plants in the garden
		plants, err := s.repos.Plant().GetByGardenID(txCtx, id)
		if err != nil {
			return err
		}

		// Delete all plants
		for _, plant := range plants {
			if err := s.repos.Plant().Delete(txCtx, plant.ID); err != nil {
				return err
			}
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
