package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"medicfo/app/model"
)

const (
	defaultDoctorLimit = 10
	maxDoctorLimit     = 50
)

func ValidateCreateDoctor(req model.CreateDoctorRequest) map[string]string {
	errs := make(map[string]string)

	if req.UserID <= 0 {
		errs["userId"] = "userId wajib diisi"
	}
	if msg := checkDoctorName(req.Name); msg != "" {
		errs["name"] = msg
	}
	if msg := checkSpecialization(req.Specialization); msg != "" {
		errs["specialization"] = msg
	}
	if msg := checkPhone(req.Phone); msg != "" {
		errs["phone"] = msg
	}
	return errs
}

func ValidateUpdateDoctor(req model.UpdateDoctorRequest) map[string]string {
	errs := make(map[string]string)

	if req.Name == nil && req.Specialization == nil && req.Phone == nil {
		errs["body"] = "minimal satu field harus diisi"
		return errs
	}
	if req.Name != nil {
		if msg := checkDoctorName(*req.Name); msg != "" {
			errs["name"] = msg
		}
	}
	if req.Specialization != nil {
		if msg := checkSpecialization(*req.Specialization); msg != "" {
			errs["specialization"] = msg
		}
	}
	if req.Phone != nil {
		if msg := checkPhone(*req.Phone); msg != "" {
			errs["phone"] = msg
		}
	}
	return errs
}

func checkDoctorName(name string) string {
	switch n := strings.TrimSpace(name); {
	case n == "":
		return "name wajib diisi"
	case len(n) > 100:
		return "name maksimal 100 karakter"
	}
	return ""
}

func checkSpecialization(s string) string {
	switch v := strings.TrimSpace(s); {
	case v == "":
		return "specialization wajib diisi"
	case len(v) > 100:
		return "specialization maksimal 100 karakter"
	}
	return ""
}

func checkPhone(phone string) string {
	p := strings.TrimSpace(phone)
	if p == "" {
		return "phone wajib diisi"
	}
	if len(p) < 8 || len(p) > 20 {
		return "phone harus 8-20 karakter"
	}
	for i, r := range p {
		if r >= '0' && r <= '9' {
			continue
		}
		if i == 0 && r == '+' {
			continue
		}
		return "phone hanya boleh angka (boleh diawali +)"
	}
	return ""
}

func NormalizeDoctorLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultDoctorLimit
	case limit > maxDoctorLimit:
		return maxDoctorLimit
	}
	return limit
}

func EncodeDoctorCursor(createdAt time.Time, id int64) string {
	raw := fmt.Sprintf("%d|%d", createdAt.UnixMicro(), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeDoctorCursor(cursor string) (time.Time, int64, error) {
	raw, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, errors.New("cursor tidak valid")
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 2 {
		return time.Time{}, 0, errors.New("cursor tidak valid")
	}
	micro, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, 0, errors.New("cursor tidak valid")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return time.Time{}, 0, errors.New("cursor tidak valid")
	}
	return time.UnixMicro(micro).UTC(), id, nil
}
