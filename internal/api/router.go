package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"checkin.service/internal/api/handler"
	checkin_service "checkin.service/internal/core/service"
)

// NewRouter sets up the gorilla/mux router and defines all API routes.
func NewRouter(service checkin_service.CheckInService) *mux.Router {

	checkInHandler := handler.CheckInHandler{
		Service: service,
	}

	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/checkin-checkout", checkInHandler.CheckInOut).Methods(http.MethodPost)
	api.HandleFunc("/checkin/{employeeId}", checkInHandler.GetCheckIn).Methods(http.MethodPost)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is operational."))
	}).Methods(http.MethodGet)

	return r
}
