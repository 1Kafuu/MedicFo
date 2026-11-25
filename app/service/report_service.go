package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/helper"
)

type ReportService struct {
	reports repository.ReportRepository
}

func NewReportService(reports repository.ReportRepository) *ReportService {
	return &ReportService{reports: reports}
}

// AppointmentsByDate memvalidasi tanggal sebelum query. Format ditetapkan
// di sini, bukan di route, karena ini aturan bisnis laporan.
func (s *ReportService) AppointmentsByDate(ctx context.Context, date string) ([]model.AppointmentReport, error) {
	date = strings.TrimSpace(date)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, &helper.AppError{
			Code:   "VALIDATION_ERROR",
			Status: 400,
			Detail: "date wajib berformat YYYY-MM-DD",
		}
	}

	items, err := s.reports.ListAppointmentsByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("laporan appointments: %w", err)
	}
	return items, nil
}
