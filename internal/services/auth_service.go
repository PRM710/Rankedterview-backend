package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PRM710/Rankedterview-backend/internal/config"
	"github.com/PRM710/Rankedterview-backend/internal/models"
	"github.com/PRM710/Rankedterview-backend/internal/repositories"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	userRepo *repositories.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// RegisterWithOAuth registers or logs in a user via OAuth
func (s *AuthService) RegisterWithOAuth(ctx context.Context, provider, oauthID, email, name, avatar string) (*models.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByOAuthID(ctx, provider, oauthID)

	if err == nil {
		// User exists, update last login and return token
		s.userRepo.UpdateLastLogin(ctx, existingUser.ID.Hex())
		token, err := s.generateToken(existingUser)
		if err != nil {
			return nil, "", err
		}
		return existingUser, token, nil
	}

	if err != mongo.ErrNoDocuments {
		// Real error occurred
		return nil, "", err
	}

	// User doesn't exist, create new user
	user := &models.User{
		Email:         email,
		Name:          name,
		Avatar:        avatar,
		OAuthProvider: provider,
		OAuthID:       oauthID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login attempts to log in a user (for future email/password auth)
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	// TODO: Implement email/password auth when needed
	return nil, "", errors.New("email/password auth not implemented")
}

// RefreshToken generates a new access token
func (s *AuthService) RefreshToken(ctx context.Context, userID string) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", ErrUserNotFound
	}

	return s.generateToken(user)
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(tokenString, s.config.JWTSecret)
}

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	expiration, err := utils.ParseDuration(s.config.JWTExpiration)
	if err != nil {
		expiration = 15 * time.Minute // Default to 15 minutes
	}

	return utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		s.config.JWTSecret,
		expiration,
	)
}

// GetOAuthURL generates the OAuth URL for a provider
func (s *AuthService) GetOAuthURL(provider string) (string, error) {
	switch provider {
	case "google":
		return fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile&access_type=offline",
			s.config.GoogleClientID,
			url.QueryEscape(s.config.GoogleRedirectURI),
		), nil
	case "github":
		return fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
			s.config.GitHubClientID,
			url.QueryEscape(s.config.GitHubRedirectURI),
		), nil
	default:
		return "", errors.New("unsupported OAuth provider")
	}
}
