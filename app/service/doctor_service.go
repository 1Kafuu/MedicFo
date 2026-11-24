package service

import (
	"context"
	"errors"
	"strings"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/helper"
)

type DoctorService struct {
	doctors repository.DoctorRepository
	users   repository.UserRepository
}

func NewDoctorService(doctors repository.DoctorRepository, users repository.UserRepository) *DoctorService {
	return &DoctorService{doctors: doctors, users: users}
}

func (s *DoctorService) Create(ctx context.Context, req model.CreateDoctorRequest) (model.Doctor, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Specialization = strings.TrimSpace(req.Specialization)
	req.Phone = strings.TrimSpace(req.Phone)

	if errs := ValidateCreateDoctor(req); len(errs) > 0 {
		return model.Doctor{}, validationError(errs)
	}

	user, err := s.users.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Doctor{}, validationError(map[string]string{"userId": "user tidak ditemukan"})
		}
		return model.Doctor{}, helper.ErrInternal
	}
	if user.Role != "doctor" {
		return model.Doctor{}, validationError(map[string]string{"userId": "user bukan ber-role doctor"})
	}

	created, err := s.doctors.Create(ctx, model.Doctor{
		UserID:         req.UserID,
		Name:           req.Name,
		Specialization: req.Specialization,
		Phone:          req.Phone,
	})
	if err != nil {
		return model.Doctor{}, translateDoctorError(err)
	}
	return created, nil
}

func (s *DoctorService) Update(ctx context.Context, id int64, req model.UpdateDoctorRequest) (model.Doctor, error) {
	trimPtr(req.Name)
	trimPtr(req.Specialization)
	trimPtr(req.Phone)

	if errs := ValidateUpdateDoctor(req); len(errs) > 0 {
		return model.Doctor{}, validationError(errs)
	}

	updated, err := s.doctors.Update(ctx, id, req)
	if err != nil {
		return model.Doctor{}, translateDoctorError(err)
	}
	return updated, nil
}

func (s *DoctorService) Get(ctx context.Context, id int64) (model.Doctor, error) {
	d, err := s.doctors.FindByID(ctx, id)
	if err != nil {
		return model.Doctor{}, translateDoctorError(err)
	}
	return d, nil
}

func (s *DoctorService) List(ctx context.Context, limit int, cursor string) (model.DoctorPage, error) {
	limit = NormalizeDoctorLimit(limit)

	var (
		rows []model.Doctor
		err  error
	)
	if cursor == "" {
		rows, err = s.doctors.List(ctx, limit+1, nil, 0)
	} else {
		createdAt, id, decodeErr := DecodeDoctorCursor(cursor)
		if decodeErr != nil {
			return model.DoctorPage{}, validationError(map[string]string{"cursor": decodeErr.Error()})
		}
		rows, err = s.doctors.List(ctx, limit+1, &createdAt, id)
	}
	if err != nil {
		return model.DoctorPage{}, helper.ErrInternal
	}

	page := model.DoctorPage{Items: rows}
	if len(rows) > limit {
		page.HasMore = true
		page.Items = rows[:limit]
		last := page.Items[limit-1]
		page.NextCursor = EncodeDoctorCursor(last.CreatedAt, last.ID)
	}
	return page, nil
}

func trimPtr(p *string) {
	if p != nil {
		*p = strings.TrimSpace(*p)
	}
}

func translateDoctorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.ErrNotFound
	case errors.Is(err, repository.ErrDuplicate):
		return helper.ErrDuplicateDoctor
	default:
		return helper.ErrInternal
	}
}
