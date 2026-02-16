package users

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gsupert/internal/config"
	"gsupert/internal/modules/auth"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
	cfg  *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) Login(email, password string) (*auth.TokenPair, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Role, s.cfg)
	if err != nil {
		return nil, err
	}

	// Store refresh token hash
	hash := sha256.Sum256([]byte(tokens.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	rt := &RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().AddDate(0, 0, s.cfg.RefreshTokenExp),
	}

	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	claims, err := auth.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	rt, err := s.repo.FindRefreshToken(tokenHash)
	if err != nil {
		return nil, errors.New("refresh token not found or revoked")
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new pair
	tokens, err := auth.GenerateTokenPair(user.ID, user.Role, s.cfg)
	if err != nil {
		return nil, err
	}

	// Remove old and store new refresh token
	s.repo.DeleteRefreshToken(user.ID)

	newHash := sha256.Sum256([]byte(tokens.RefreshToken))
	newTokenHash := hex.EncodeToString(newHash[:])

	newRt := &RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().AddDate(0, 0, s.cfg.RefreshTokenExp),
	}

	if err := s.repo.CreateRefreshToken(newRt); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) Logout(userID string) error {
	return s.repo.DeleteRefreshToken(userID)
}

func (s *Service) CreateUser(email, password, fullName, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		Role:         role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetUser(id string) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListUsers() ([]User, error) {
	return s.repo.List()
}

func (s *Service) UpdateUser(id string, email, fullName, role string) (*User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.Email = email
	user.FullName = fullName
	user.Role = role

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) DeleteUser(id string) error {
	return s.repo.Delete(id)
}
