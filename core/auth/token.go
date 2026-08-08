package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"uid"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	ExpiresAt int64     `json:"exp"`
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *TokenService) Issue(userID uuid.UUID, email string, roles []string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		ExpiresAt: time.Now().Add(s.ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + s.sign(body), nil
}

func (s *TokenService) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}
	body, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sig), []byte(s.sign(body))) {
		return nil, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

func (s *TokenService) sign(body string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
