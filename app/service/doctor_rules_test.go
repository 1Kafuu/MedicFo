package service

import (
	"testing"
	"time"

	"medicfo/app/model"
)

func TestDoctorCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 11, 24, 10, 30, 0, 123456000, time.UTC)

	cursor := EncodeDoctorCursor(createdAt, 42)
	gotTime, gotID, err := DecodeDoctorCursor(cursor)
	if err != nil {
		t.Fatalf("decode gagal: %v", err)
	}
	if !gotTime.Equal(createdAt) || gotID != 42 {
		t.Fatalf("hasil %v/%d, harapan %v/42", gotTime, gotID, createdAt)
	}
}

func TestDecodeDoctorCursorRejectsGarbage(t *testing.T) {
	for _, bad := range []string{"%%%", "YWJj", "MXwwfDA=", "MXxhYmM="} {
		if _, _, err := DecodeDoctorCursor(bad); err == nil {
			t.Errorf("cursor %q seharusnya ditolak", bad)
		}
	}
}

func TestNormalizeDoctorLimit(t *testing.T) {
	cases := map[int]int{-5: 10, 0: 10, 7: 7, 50: 50, 500: 50}
	for in, want := range cases {
		if got := NormalizeDoctorLimit(in); got != want {
			t.Errorf("limit %d: dapat %d, harapan %d", in, got, want)
		}
	}
}

func TestValidateCreateDoctor(t *testing.T) {
	valid := model.CreateDoctorRequest{UserID: 3, Name: "dr. Sinta", Specialization: "Cardiology", Phone: "08121234567"}
	if errs := ValidateCreateDoctor(valid); len(errs) != 0 {
		t.Fatalf("request valid ditolak: %v", errs)
	}

	invalid := model.CreateDoctorRequest{Phone: "abc"}
	errs := ValidateCreateDoctor(invalid)
	for _, field := range []string{"userId", "name", "specialization", "phone"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("field %s seharusnya error", field)
		}
	}
}

func TestValidateUpdateDoctorNeedsOneField(t *testing.T) {
	if errs := ValidateUpdateDoctor(model.UpdateDoctorRequest{}); len(errs) == 0 {
		t.Fatal("body kosong seharusnya ditolak")
	}
	name := "dr. Budi"
	if errs := ValidateUpdateDoctor(model.UpdateDoctorRequest{Name: &name}); len(errs) != 0 {
		t.Fatalf("update valid ditolak: %v", errs)
	}
}
