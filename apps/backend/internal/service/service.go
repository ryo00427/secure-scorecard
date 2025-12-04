package service

import (
	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// Service provides business logic
type Service struct {
	repo *repository.Repository
}

// NewService creates a new Service instance
func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// --- User Service Methods ---

// CreateUser creates a new user
func (s *Service) CreateUser(user *model.User) error {
	return s.repo.CreateUser(user)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

// GetUserByFirebaseUID retrieves a user by Firebase UID
func (s *Service) GetUserByFirebaseUID(uid string) (*model.User, error) {
	return s.repo.GetUserByFirebaseUID(uid)
}

// GetOrCreateUser gets an existing user or creates a new one
func (s *Service) GetOrCreateUser(firebaseUID, email, displayName, photoURL string) (*model.User, error) {
	user, err := s.repo.GetUserByFirebaseUID(firebaseUID)
	if err == nil {
		return user, nil
	}

	// Create new user
	newUser := &model.User{
		FirebaseUID: firebaseUID,
		Email:       email,
		DisplayName: displayName,
		PhotoURL:    photoURL,
		IsActive:    true,
	}

	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// --- Garden Service Methods ---

// CreateGarden creates a new garden for a user
func (s *Service) CreateGarden(userID uint, name, description, location string, sizeM2 float64) (*model.Garden, error) {
	garden := &model.Garden{
		UserID:      userID,
		Name:        name,
		Description: description,
		Location:    location,
		SizeM2:      sizeM2,
	}

	if err := s.repo.CreateGarden(garden); err != nil {
		return nil, err
	}

	return garden, nil
}

// GetGardenByID retrieves a garden by ID
func (s *Service) GetGardenByID(id uint) (*model.Garden, error) {
	return s.repo.GetGardenByID(id)
}

// GetUserGardens retrieves all gardens for a user
func (s *Service) GetUserGardens(userID uint) ([]model.Garden, error) {
	return s.repo.GetGardensByUserID(userID)
}

// UpdateGarden updates a garden
func (s *Service) UpdateGarden(garden *model.Garden) error {
	return s.repo.UpdateGarden(garden)
}

// DeleteGarden soft deletes a garden
func (s *Service) DeleteGarden(id uint) error {
	return s.repo.DeleteGarden(id)
}

// --- Plant Service Methods ---

// CreatePlant creates a new plant in a garden
func (s *Service) CreatePlant(plant *model.Plant) error {
	return s.repo.CreatePlant(plant)
}

// GetPlantByID retrieves a plant by ID
func (s *Service) GetPlantByID(id uint) (*model.Plant, error) {
	return s.repo.GetPlantByID(id)
}

// GetGardenPlants retrieves all plants in a garden
func (s *Service) GetGardenPlants(gardenID uint) ([]model.Plant, error) {
	return s.repo.GetPlantsByGardenID(gardenID)
}

// UpdatePlant updates a plant
func (s *Service) UpdatePlant(plant *model.Plant) error {
	return s.repo.UpdatePlant(plant)
}

// DeletePlant soft deletes a plant
func (s *Service) DeletePlant(id uint) error {
	return s.repo.DeletePlant(id)
}

// --- CareLog Service Methods ---

// CreateCareLog creates a new care log for a plant
func (s *Service) CreateCareLog(careLog *model.CareLog) error {
	return s.repo.CreateCareLog(careLog)
}

// GetPlantCareLogs retrieves all care logs for a plant
func (s *Service) GetPlantCareLogs(plantID uint) ([]model.CareLog, error) {
	return s.repo.GetCareLogsByPlantID(plantID)
}
