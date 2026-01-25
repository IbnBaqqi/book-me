package main

import (
	"context"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
)

	type ReservedSlotDto struct {
		ID        int64     `json:"id"`
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
		BookedBy  *string   `json:"bookedBy,omitempty"`
	}

	type ReservedDto struct {
		RoomID   int64             `json:"roomId"`
		RoomName string            `json:"roomName"`
		Slots    []ReservedSlotDto `json:"slots"`
	}

func (cfg *apiConfig) handlerFetchReservations(w http.ResponseWriter, r *http.Request) {

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start")
	endDateStr := r.URL.Query().Get("end")

	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "start and end parameters are required", nil)
		return
	}

	// Parse dates (ISO format: YYYY-MM-DD)
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid end date format", err)
		return
	}

	// Get authenticated user from context (can be nil for public access)
	currentUser, isAuthenticated := auth.UserFromContext(r.Context())

	// Call service method
	reserved, err := cfg.getUnavailableSlots(r.Context(), startDate, endDate, currentUser, isAuthenticated)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to fetch unavailable slots", err)
		return
	}

	respondWithJSON(w, http.StatusOK, reserved)

}

// Service method
func (cfg *apiConfig) getUnavailableSlots(
	ctx context.Context,
	start time.Time,
	end time.Time,
	currentUser auth.User,
	isAuthenticated bool,
) ([]ReservedDto, error) {

	// Convert dates to datetime range
	startDateTime := start // Already at 00:00:00
	endDateTime := end.AddDate(0, 0, 1) // Add 1 day (equivalent to plusDays(1).atStartOfDay())

	// Check if user is staff
	isStaff := isAuthenticated && currentUser.Role == "STAFF"

	// Fetch all reservations between dates
	reservations, err := cfg.db.GetAllBetweenDates(ctx, database.GetAllBetweenDatesParams{
		StartTime: startDateTime,
		EndTime:   endDateTime,
	})
	if err != nil {
		return nil, err // handle db error
	}
	// Group reservations by room ID
	grouped := make(map[int64][]database.GetAllBetweenDatesRow)
	for _, res := range reservations {
		grouped[res.RoomID] = append(grouped[res.RoomID], res)
	}

	// Build result
	result := make([]ReservedDto, 0, len(grouped))

	for roomID, roomReservations := range grouped {
		if len(roomReservations) == 0 {
			continue
		}

		roomName := roomReservations[0].RoomName

		// Map reservations to slots
		slots := make([]ReservedSlotDto, 0, len(roomReservations))
		for _, res := range roomReservations {
			var bookedBy *string

			// Show bookedBy only if user is staff or is the owner
			if isStaff || (isAuthenticated && res.CreatedByID == int64(currentUser.ID)) {
				bookedBy = &res.CreatedByName
			}

			slots = append(slots, ReservedSlotDto{
				ID:        res.ID,
				StartTime: res.StartTime,
				EndTime:   res.EndTime,
				BookedBy:  bookedBy,
			})
		}

		result = append(result, ReservedDto{
			RoomID:   roomID,
			RoomName: roomName,
			Slots:    slots,
		})
	}

	return result, nil
}