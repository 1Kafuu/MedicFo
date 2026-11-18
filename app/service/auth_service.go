package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/helper"
)

// Cukup panjang agar tidak bisa ditebak; tidak perlu lebih panjang.
const refreshTokenBytes = 32


type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(users repository.UserRepository, tokens repository.TokenRepository,
	jwtManager *helper.JWTManager, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwt: jwtManager, refreshTTL: refreshTTL}
}

// Register membuat akun baru dengan role patient.
// Role TIDAK pernah dibaca dari request (mass assignment).
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.User, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := ValidateRegister(req); len(errs) > 0 {
		return model.User{}, validationError(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return model.User{}, helper.ErrInternal
	}

	created, err := s.users.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "patient",
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return model.User{}, helper.ErrDuplicateUser
		}
		return model.User{}, helper.ErrInternal
	}
	return created, nil
}

// Login memeriksa kredensial lalu menerbitkan pasangan token.
// Email tidak dikenal dibalas persis seperti password salah,
// lengkap dengan VerifyDummyPassword agar pola waktunya serupa.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.TokenPair, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := ValidateLogin(req); len(errs) > 0 {
		return model.TokenPair{}, validationError(errs)
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		return model.TokenPair{}, helper.ErrInvalidCredentials
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return model.TokenPair{}, helper.ErrInvalidCredentials
	}

	if !user.IsActive {
		return model.TokenPair{}, helper.ErrForbidden
	}

	return s.issueTokenPair(ctx, user)
}

// Refresh memutar token: refresh token lama dicabut, pasangan baru diterbitkan.
func (s *AuthService) Refresh(ctx context.Context, req model.RefreshRequest) (model.TokenPair, error) {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return model.TokenPair{}, helper.ErrBadRequest
	}

	hash := helper.SHA256Hex(req.RefreshToken)

	stored, err := s.tokens.FindActive(ctx, hash)
	if err != nil {
		// Token tak dikenal / dicabut / kedaluwarsa → satu pesan yang sama.
		return model.TokenPair{}, helper.ErrUnauthorized
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil || !user.IsActive {
		return model.TokenPair{}, helper.ErrUnauthorized
	}

	// Rotasi: token lama mati SEBELUM token baru diterbitkan.
	if err := s.tokens.Revoke(ctx, hash); err != nil {
		return model.TokenPair{}, helper.ErrInternal
	}

	return s.issueTokenPair(ctx, user)
}

// Logout mencabut refresh token.
func (s *AuthService) Logout(ctx context.Context, req model.RefreshRequest) error {
	if strings.TrimSpace(req.RefreshToken) != "" {
		_ = s.tokens.Revoke(ctx, helper.SHA256Hex(req.RefreshToken))
	}
	return nil
}

// Me mengembalikan profil user yang sedang login (dari access token).
func (s *AuthService) Me(ctx context.Context, authUser model.AuthUser) (model.User, error) {
	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return model.User{}, helper.ErrUnauthorized
	}
	return user, nil
}

// issueTokenPair menerbitkan access token + refresh token baru
// dan menyimpan hash refresh token.
func (s *AuthService) issueTokenPair(ctx context.Context, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, helper.ErrInternal
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, helper.ErrInternal
	}

	if err := s.tokens.Save(ctx, user.ID, helper.SHA256Hex(refreshToken), time.Now().Add(s.refreshTTL)); err != nil {
		return model.TokenPair{}, helper.ErrInternal
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}
