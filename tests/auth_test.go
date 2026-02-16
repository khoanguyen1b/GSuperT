package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user/gsupert/internal/config"
	"github.com/user/gsupert/internal/modules/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashing(t *testing.T) {
	password := "abcd@123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestJWTGenerationAndVerification(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:       "test_secret",
		JWTRefreshSecret: "test_refresh_secret",
		AccessTokenExp:  60,
		RefreshTokenExp: 7,
	}

	userID := "user-123"
	role := "admin"

	tokens, err := auth.GenerateTokenPair(userID, role, cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Verify Access Token
	claims, err := auth.ValidateToken(tokens.AccessToken, cfg.JWTSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)

	// Verify Refresh Token
	refreshClaims, err := auth.ValidateToken(tokens.RefreshToken, cfg.JWTRefreshSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)

	// Test Invalid Token
	_, err = auth.ValidateToken("invalid.token.string", cfg.JWTSecret)
	assert.Error(t, err)
}
