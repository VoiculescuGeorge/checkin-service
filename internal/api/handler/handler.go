package handler

import (
	"encoding/json"
	"net/http"

	checkin_service "checkin.service/internal/core"
)

type CheckInHandler struct {
	Service checkin_service.CheckInService
}

type CheckInOutRequest struct {
	EmployeeID string `json:"employeeId"`
}

func (h *CheckInHandler) CheckInOut(w http.ResponseWriter, r *http.Request) {
	var req CheckInOutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.EmployeeID == "" {
		http.Error(w, "EmployeeID is required", http.StatusBadRequest)
		return
	}

	err := h.Service.ProcessCheckInOut(r.Context(), req.EmployeeID)

	if err != nil {
		http.Error(w, "Service error processing event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{"message": "Check-in/out event recorded for asynchronous processing."})
}
