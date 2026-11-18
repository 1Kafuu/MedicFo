package service

import (
	"sort"
	"strings"
	"unicode"

	"medicfo/app/model"
	"medicfo/helper"
)

// validateRegister memeriksa DTO register.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := make(map[string]string)

	switch u := strings.TrimSpace(req.Username); {
	case u == "":
		errs["username"] = "username wajib diisi"
	case len(u) < 3:
		errs["username"] = "username minimal 3 karakter"
	case !isValidUsername(u):
		errs["username"] = "hanya boleh huruf, angka, titik, dan underscore"
	}

	switch e := strings.TrimSpace(req.Email); {
	case e == "":
		errs["email"] = "email wajib diisi"
	case !isValidEmail(e):
		errs["email"] = "format email tidak valid"
	}

	if msg := checkPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}

	return errs
}

// validateLogin memeriksa DTO login (email + password wajib ada).
func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := make(map[string]string)

	if strings.TrimSpace(req.Email) == "" {
		errs["email"] = "email wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "password wajib diisi"
	}

	return errs
}

func isValidUsername(u string) bool {
	for _, r := range u {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
			return false
		}
	}
	return true
}

func isValidEmail(e string) bool {
	at := strings.Index(e, "@")
	if at <= 0 || at == len(e)-1 || strings.Count(e, "@") != 1 {
		return false
	}
	domain := e[at+1:]
	return strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}

// checkPasswordStrength mengembalikan "" bila kuat, selain itu pesannya.
func checkPasswordStrength(p string) string {
	if len(p) < 8 {
		return "password minimal 8 karakter"
	}

	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "password harus mengandung huruf dan angka"
	}

	switch strings.ToLower(p) {
	case "password1", "12345678", "qwerty123", "admin123", "password123":
		return "password terlalu mudah ditebak"
	}
	return ""
}

func validationError(errs map[string]string) error {
	fields := make([]string, 0, len(errs))
	for field, msg := range errs {
		fields = append(fields, field+": "+msg)
	}
	sort.Strings(fields)

	return &helper.AppError{
		Code:   "VALIDATION_ERROR",
		Status: 400,
		Detail: strings.Join(fields, "; "),
	}
}
