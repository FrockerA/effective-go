package handler

import (
	_ "effective-go/internal/model"
	"effective-go/internal/service"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

// Create godoc
// @Summary Создать подписку
// @Description Добавляет новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body service.CreateDTO true "Данные подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /subscriptions/ [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.service.Create(r.Context(), dto)
	if err != nil {

		if errors.Is(err, service.ErrInvalidDateFormat) || errors.Is(err, service.ErrInvalidDateOrder) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// List godoc
// @Summary Получить список подписок
// @Description Возвращает список подписок с возможностью фильтрации по user_id и service_name
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Success 200 {array} model.Subscription
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /subscriptions/ [get]
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

// GetByID godoc
// @Summary Получить подписку по ID
// @Description Возвращает полные данные подписки по её уникальному идентификатору
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// Update godoc
// @Summary Обновить подписку
// @Description Обновляет данные существующей подписки (название, цену, даты)
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Param input body service.UpdateDTO true "Новые данные подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	var dto service.UpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.service.Update(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidDateFormat) || errors.Is(err, service.ErrInvalidDateOrder) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update subscription")
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// Delete godoc
// @Summary Удалить подписку
// @Description Полностью удаляет запись о подписке из базы данных
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		writeError(w, http.StatusBadRequest, "start_date and end_date query parameters are required")
		return
	}

	total, err := h.service.CalculateTotal(r.Context(), userID, serviceName, startDateStr, endDateStr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateFormat) || errors.Is(err, service.ErrInvalidDateOrder) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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
