package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type tokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	TokenID   string `json:"jti"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

func NewTokenManager(cfg Config) (*TokenManager, error) {
	if strings.TrimSpace(cfg.AccessSecret) == "" || strings.TrimSpace(cfg.RefreshSecret) == "" {
		return nil, fmt.Errorf("jwt secrets are required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, fmt.Errorf("jwt ttl values must be positive: %w", coreerrors.ErrInvalidDomainValue)
	}

	return &TokenManager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}, nil
}

func (m *TokenManager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *TokenManager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func (m *TokenManager) GenerateAccessToken(user users.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.accessTTL)
	token, err := m.generateToken(user, TokenTypeAccess, expiresAt, m.accessSecret)
	return token, expiresAt, err
}

func (m *TokenManager) GenerateRefreshToken(user users.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.refreshTTL)
	token, err := m.generateToken(user, TokenTypeRefresh, expiresAt, m.refreshSecret)
	return token, expiresAt, err
}

func (m *TokenManager) ParseAccessToken(raw string) (Claims, error) {
	return m.parseToken(raw, TokenTypeAccess, m.accessSecret)
}

func (m *TokenManager) ParseRefreshToken(raw string) (Claims, error) {
	return m.parseToken(raw, TokenTypeRefresh, m.refreshSecret)
}

func (m *TokenManager) generateToken(user users.User, tokenType TokenType, expiresAt time.Time, secret []byte) (string, error) {
	issuedAt := time.Now().UTC()
	tokenID, err := GenerateOpaqueToken()
	if err != nil {
		return "", err
	}

	claims := tokenClaims{
		UserID:    user.ID.String(),
		Email:     user.Email,
		Role:      user.Role.String(),
		TokenType: string(tokenType),
		TokenID:   tokenID,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  issuedAt.Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := encodedHeader + "." + encodedClaims
	signature := signHS256(signingInput, secret)

	return signingInput + "." + signature, nil
}

func (m *TokenManager) parseToken(raw string, expectedType TokenType, secret []byte) (Claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Claims{}, fmt.Errorf("invalid token format: %w", coreerrors.ErrUnauthorized)
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signHS256(signingInput, secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, fmt.Errorf("invalid token signature: %w", coreerrors.ErrUnauthorized)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, fmt.Errorf("decode token claims: %w", coreerrors.ErrUnauthorized)
	}

	var rawClaims tokenClaims
	if err := json.Unmarshal(payload, &rawClaims); err != nil {
		return Claims{}, fmt.Errorf("unmarshal token claims: %w", coreerrors.ErrUnauthorized)
	}

	if rawClaims.TokenType != string(expectedType) {
		return Claims{}, fmt.Errorf("unexpected token type: %w", coreerrors.ErrUnauthorized)
	}
	if time.Now().UTC().Unix() >= rawClaims.ExpiresAt {
		return Claims{}, fmt.Errorf("token expired: %w", coreerrors.ErrUnauthorized)
	}

	userID, err := uuid.Parse(rawClaims.UserID)
	if err != nil {
		return Claims{}, fmt.Errorf("parse user id claim: %w", coreerrors.ErrUnauthorized)
	}
	role, err := users.ParseRole(rawClaims.Role)
	if err != nil {
		return Claims{}, fmt.Errorf("parse role claim: %w", coreerrors.ErrUnauthorized)
	}

	return Claims{
		UserID:    userID,
		Email:     rawClaims.Email,
		Role:      role,
		TokenType: expectedType,
		ExpiresAt: time.Unix(rawClaims.ExpiresAt, 0).UTC(),
		IssuedAt:  time.Unix(rawClaims.IssuedAt, 0).UTC(),
	}, nil
}

func GenerateOpaqueToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate opaque token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func signHS256(input string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
