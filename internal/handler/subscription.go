package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"effective-go/internal/service"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.service.Create(r.Context(), dto)
	if err != nil {
		// Обработка бизнес-ошибок валидации
		if errors.Is(err, service.ErrInvalidDateFormat) || errors.Is(err, service.ErrInvalidDateOrder) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	uidStr := r.URL.Query().Get("user_id")
	if uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id format")
			return
		}
		userID = &uid
	}

	serviceName := r.URL.Query().Get("service_name")

	subs, err := h.service.List(r.Context(), userID, serviceName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch subscriptions")
		return
	}

	if subs == nil {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

func (h *SubscriptionHandler) CalculateTotal(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	uidStr := r.URL.Query().Get("user_id")
	if uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id format")
			return
		}
		userID = &uid
	}

	serviceName := r.URL.Query().Get("service_name")

	total, err := h.service.CalculateTotal(r.Context(), userID, serviceName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate total")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"total_cost": total})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
