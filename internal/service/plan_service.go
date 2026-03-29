package service

import (
	"encoding/json"
	"net/http"
	"time"

	"dailyApp/backend/internal/model"

	"gorm.io/gorm"
)

type PlanService struct {
	db *gorm.DB
}

type planResponse struct {
	ID             string    `json:"id"`
	ActivityID     string    `json:"activityId"`
	Date           string    `json:"date"`
	StartMinute    int       `json:"startMinute"`
	PlannedMinutes int       `json:"plannedMinutes"`
	Memo           string    `json:"memo"`
	CreatedAt      time.Time `json:"createdAt"`
}

func toPlanResponse(p model.DailyPlan) planResponse {
	return planResponse{
		ID:             p.ID,
		ActivityID:     p.ActivityID,
		Date:           p.Date,
		StartMinute:    p.StartMinute,
		PlannedMinutes: p.PlannedMinutes,
		Memo:           p.Memo,
		CreatedAt:      p.CreatedAt,
	}
}

func NewPlanService(db *gorm.DB) *PlanService {
	return &PlanService{db: db}
}

func (s *PlanService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.listPlans(w, r)
	case http.MethodPost:
		s.createPlan(w, r)
	case http.MethodDelete:
		s.deletePlan(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PlanService) listPlans(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "missing date query param", http.StatusBadRequest)
		return
	}

	var plans []model.DailyPlan
	if err := s.db.Where("date = ?", date).Order("created_at asc").Find(&plans).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]planResponse, 0, len(plans))
	for _, p := range plans {
		resp = append(resp, toPlanResponse(p))
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *PlanService) createPlan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ActivityID     string `json:"activityId"`
		Date           string `json:"date"`
		StartMinute    int    `json:"startMinute"`
		PlannedMinutes int    `json:"plannedMinutes"`
		Memo           string `json:"memo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan := model.DailyPlan{
		ActivityID:     body.ActivityID,
		Date:           body.Date,
		StartMinute:    body.StartMinute,
		PlannedMinutes: body.PlannedMinutes,
		Memo:           body.Memo,
	}
	if err := s.db.Create(&plan).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toPlanResponse(plan))
}

func (s *PlanService) deletePlan(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id query param", http.StatusBadRequest)
		return
	}

	if err := s.db.Delete(&model.DailyPlan{}, "id = ?", id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
