package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// A simple struct to capture the incoming event data
type CheckOutEvent struct {
	WorkingTimeID int64   `json:"working_time_id"`
	EmployeeID    string  `json:"employee_id"`
	HoursWorked   float64 `json:"hours_worked"`
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	var event CheckOutEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Received checkout for EmployeeID: %s, Hours: %.2f", event.EmployeeID, event.HoursWorked)
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", checkoutHandler)
	log.Println("Legacy API mock server starting on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
