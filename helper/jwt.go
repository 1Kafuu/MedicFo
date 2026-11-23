package helper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"medicfo/app/model"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

// accessClaims: isi access token. sub = user id; username & role dibawa
// agar middleware tidak perlu query database pada setiap request.
type accessClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

// NewJWTManager menerima secret dari luar (main.go)
func NewJWTManager(secret, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

// AccessTTL mengembalikan masa berlaku access token (untuk field expiresIn).
func (m *JWTManager) AccessTTL() time.Duration { return m.accessTTL }

// GenerateAccess membuat access token berumur pendek.
func (m *JWTManager) GenerateAccess(u model.User) (string, error) {
	now := time.Now()

	claims := accessClaims{
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse memeriksa tanda tangan, issuer, dan masa berlaku, lalu
// mengembalikan identitas yang dibawa token.
func (m *JWTManager) Parse(tokenString string) (model.AuthUser, error) {
	claims := &accessClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritma tidak diharapkan: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.AuthUser{}, ErrExpiredToken
		}
		return model.AuthUser{}, ErrInvalidToken
	}

	if !token.Valid {
		return model.AuthUser{}, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}

	return model.AuthUser{UserID: userID, Username: claims.Username, Role: claims.Role}, nil
}
