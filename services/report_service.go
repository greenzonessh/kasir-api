package services

import (
	"context"
	"time"

	"kasir-api/models"
	"kasir-api/repositories"
)

type ReportService struct {
	repo *repositories.ReportRepository
}

func NewReportService(repo *repositories.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

// ReportHariIni sesuai zona Asia/Jakarta, dikonversi ke UTC untuk query.
func (s *ReportService) ReportHariIni(ctx context.Context) (models.ReportResponse, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	startLocal := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endLocal := startLocal.Add(24 * time.Hour)

	// Konversi ke UTC (umumnya efektif di Supabase/PG)
	startUTC := startLocal.In(time.UTC)
	endUTC := endLocal.In(time.UTC)

	return s.repo.FetchReport(ctx, startUTC, endUTC)
}

// ReportByRange menerima tanggal (start inclusive, end inclusive di request, kita ubah ke eksklusif)
func (s *ReportService) ReportByRange(ctx context.Context, startLocal, endLocal time.Time) (models.ReportResponse, error) {
	// endLocal dari user dianggap end-of-day inklusif; jadikan eksklusif dengan +1 hari 00:00
	endLocal = endLocal.AddDate(0, 0, 1)

	startUTC := startLocal.In(time.UTC)
	endUTC := endLocal.In(time.UTC)

	return s.repo.FetchReport(ctx, startUTC, endUTC)
}
