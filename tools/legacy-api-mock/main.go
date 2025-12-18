package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// A simple struct to capture the incoming event data
type CheckOutEvent struct {
	WorkingTimeID int64     `json:"workingTimeId"`
	EmployeeID    string    `json:"employeeId"`
	HoursWorked   float64   `json:"hoursWorked"`
	ClockOutTime  time.Time `json:"clockOutTime"`
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
