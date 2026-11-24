package model

import "time"

type Doctor struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"-"`
	Name           string    `json:"name"`
	Specialization string    `json:"specialization"`
	Phone          string    `json:"phone"`
	CreatedAt      time.Time `json:"-"`
}

type CreateDoctorRequest struct {
	UserID         int64  `json:"userId"`
	Name           string `json:"name"`
	Specialization string `json:"specialization"`
	Phone          string `json:"phone"`
}

type UpdateDoctorRequest struct {
	Name           *string `json:"name"`
	Specialization *string `json:"specialization"`
	Phone          *string `json:"phone"`
}

type DoctorPage struct {
	Items      []Doctor `json:"items"`
	NextCursor string   `json:"nextCursor,omitempty"`
	HasMore    bool     `json:"hasMore"`
}
