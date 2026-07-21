package services

import (
	"errors"

	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Authenticate(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}
	
	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	return user, nil
}

func (s *AuthService) CreateUser(username, password, email string) (*models.User, error) {
	// Check if user already exists
	existing, _ := s.userRepo.GetByUsername(username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		IsActive:     true,
	}
	
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}
