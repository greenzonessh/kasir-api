package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kasir-api/services"
)

type ReportHandler struct {
	service *services.ReportService
}

func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ReportHariIni(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /api/report/hari-ini
func (h *ReportHandler) ReportHariIni(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.ReportHariIni(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch report: %v"), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ReportHandler) HandleReportRange(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ReportRange(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func (h *ReportHandler) ReportRange(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")
	if startStr == "" || endStr == "" {
		http.Error(w, "start_date and end_date are required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		http.Error(w, "failed to load timezone", http.StatusInternalServerError)
		return
	}
	startLocal, err := time.ParseInLocation("2006-01-02", startStr, loc)
	if err != nil {
		http.Error(w, "invalid start_date (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	endLocal, err := time.ParseInLocation("2006-01-02", endStr, loc)
	if err != nil {
		http.Error(w, "invalid end_date (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	if endLocal.Before(startLocal) {
		http.Error(w, "end_date must be >= start_date", http.StatusBadRequest)
		return
	}

	resp, err := h.service.ReportByRange(r.Context(), startLocal, endLocal)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch report: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
