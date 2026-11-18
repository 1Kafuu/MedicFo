package model

import "time"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	// Sengaja TIDAK ada field Role: registrasi publik hanya membuat patient
	// (kerentanan mass assignment, Modul 5).
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// TokenPair mengikuti kontrak API.md (camelCase, tanpa tokenType).
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"` // detik
}

// RefreshToken adalah baris pada table refresh_tokens.
// Yang disimpan TokenHash (SHA-256), bukan token mentahnya.
type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// AuthUser adalah identitas yang dibawa access token dan disimpan di Locals.
type AuthUser struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
