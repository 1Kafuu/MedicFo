package model

import "time"

// User adalah baris pada table users.
// Password menyimpan hash Bcrypt dan tidak pernah di-serialize ke JSON.
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
