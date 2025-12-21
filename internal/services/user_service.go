package services

import (
	"context"
	"errors"

	"github.com/yourusername/rankedterview-backend/internal/models"
	"github.com/yourusername/rankedterview-backend/internal/repositories"
)

var (
	ErrUpdateFailed = errors.New("failed to update user")
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(ctx context.Context, userID string, input models.UpdateUserInput) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Avatar != "" {
		user.Avatar = input.Avatar
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, ErrUpdateFailed
	}

	return user, nil
}

// GetUserStats retrieves user statistics
func (s *UserService) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user.Stats, nil
}

// UpdateUserStats updates user statistics
func (s *UserService) UpdateUserStats(ctx context.Context, userID string, stats models.UserStats) error {
	return s.userRepo.UpdateStats(ctx, userID, stats)
}

// ListUsers lists all users with pagination
func (s *UserService) ListUsers(ctx context.Context, page, limit int64) ([]*models.User, int64, error) {
	skip := (page - 1) * limit
	
	users, err := s.userRepo.List(ctx, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}
