package service

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/dto"
	"github.com/IbnBaqqi/book-me/internal/email"
)

type ReservationService struct {
	db       *database.Queries
	email    *email.Service
	calendar *google.CalendarService
}

type CreateReservationInput struct {
	UserID    int64
	UserName  string
	UserEmail string
	UserRole  string
	RoomID    int64
	StartTime time.Time
	EndTime   time.Time
}

type GetReservationsInput struct {
	StartDate time.Time
	EndDate   time.Time
	UserID    int64
	UserRole  string
}

type CancelReservationInput struct {
	ID  int64
	UserID int16
	UserRole string
}

func NewReservationService(
	db *database.Queries,
	emailService *email.Service,
	calendarService *google.CalendarService,
) *ReservationService {
	return &ReservationService{
		db:       db,
		email:    emailService,
		calendar: calendarService,
	}
}

func (s *ReservationService) CreateReservation(
	ctx context.Context,
	input CreateReservationInput,
) (*database.Reservation, error) {

	// Validate time
	if input.StartTime.Before(time.Now()) {
		return nil, ErrPastTime
	}

	if !input.EndTime.After(input.StartTime) {
		return nil, ErrInvalidTimeRange
	}

	// Fetch room
	room, err := s.db.GetRoomByID(ctx, input.RoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	// Check for overlapping reservations
	overlap, err := s.db.ExistsOverlappingReservation(ctx, database.ExistsOverlappingReservationParams{
		RoomID:    input.RoomID,
		StartTime: input.EndTime,
		EndTime:   input.StartTime,
	})
	if err != nil {
		return nil, err
	}

	if overlap {
		return nil, ErrTimeSlotTaken
	}

	// Validate duration (students only)
	duration := input.EndTime.Sub(input.StartTime)
	maxDuration := 4 * time.Hour

	if duration > maxDuration && input.UserRole == "STUDENT" {
		return nil, ErrExceedsMaxDuration
	}

	// Create reservation
	reservation, err := s.db.CreateReservation(ctx, database.CreateReservationParams{
		UserID:    input.UserID,
		RoomID:    room.ID,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Status:    "RESERVED",
	})
	if err != nil {
		return nil, err
	}

	// Create Google Calendar event (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build reservation object for calendar service
		calendarReservation := &google.Reservation{
			StartTime: reservation.StartTime,
			EndTime:   reservation.EndTime,
			CreatedBy: input.UserName,
			Room:      room.Name,
		}

		eventID, err := s.calendar.CreateGoogleEvent(ctx, calendarReservation)
		if err != nil {
			log.Printf("Failed to create Google Calendar event: %v", err)
			return
		}

		// Update reservation with event ID
		if eventID != "" {
			updateErr := s.db.UpdateGoogleCalID(ctx, database.UpdateGoogleCalIDParams{
				ID:          reservation.ID,
				GcalEventID: sql.NullString{String: eventID, Valid: eventID != ""},
			})
			if updateErr != nil {
				log.Printf("Failed to update reservation with calendar event ID: %v", updateErr)
			}
		}
	}()

	// Send confirmation email (async)
	s.email.SendConfirmation(
		ctx,
		input.UserEmail,
		room.Name,
		reservation.StartTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		reservation.EndTime.Format("Monday, January 2, 2006 at 3:04 PM"),
	)

	return &reservation, nil
}

func (h *ReservationService) GetReservations(
	ctx context.Context,
	input GetReservationsInput,
) ([]dto.ReservedDto, error) {

	// Convert dates to datetime range
	startDateTime := input.StartDate              // Already at 00:00:00
	endDateTime := input.EndDate.AddDate(0, 0, 1) // Add 1 day (equivalent to plusDays(1).atStartOfDay())

	// Check if user is a staff
	isStaff := input.UserRole == "STAFF"

	// Fetch all reservations between dates
	reservations, err := h.db.GetAllBetweenDates(ctx, database.GetAllBetweenDatesParams{
		StartTime: startDateTime,
		EndTime:   endDateTime,
	})
	if err != nil {
		return nil, ErrReservationFetchFailed
	}
	// Group reservations by room ID
	grouped := make(map[int64][]database.GetAllBetweenDatesRow)
	for _, res := range reservations {
		grouped[res.RoomID] = append(grouped[res.RoomID], res)
	}

	// Build result
	result := make([]dto.ReservedDto, 0, len(grouped))

	for roomID, roomReservations := range grouped {
		if len(roomReservations) == 0 {
			continue
		}

		roomName := roomReservations[0].RoomName

		// Map reservations to slots
		slots := make([]dto.ReservedSlotDto, 0, len(roomReservations))
		for _, res := range roomReservations {
			var bookedBy *string

			// Show bookedBy only if user is staff or is the owner
			if isStaff || res.CreatedByID == input.UserID {
				bookedBy = &res.CreatedByName
			}

			slots = append(slots, dto.ReservedSlotDto{
				ID:        res.ID,
				StartTime: res.StartTime,
				EndTime:   res.EndTime,
				BookedBy:  bookedBy,
			})
		}

		result = append(result, dto.ReservedDto{
			RoomID:   roomID,
			RoomName: roomName,
			Slots:    slots,
		})
	}

	return result, nil
}

func (h *ReservationService) CancelReservation(
	ctx context.Context,
	input CancelReservationInput,
) error {

	// Find reservation by ID
	reservation, err := h.db.GetReservationByID(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReservationNotFound
		}
		return err
	}

	// Check authorization
	isStaff := input.UserRole == "STAFF"
	isOwner := reservation.UserID == int64(input.UserID)

	if !isStaff && !isOwner {
		return ErrUnauthorizedCancellation
	}

	// Delete from database
	err = h.db.DeleteReservation(ctx, input.ID)
	if err != nil {
		return err
	}

	go func ()  {
		ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
		defer cancel()
		h.calendar.DeleteGoogleEvent(ctx, reservation.GcalEventID.String)
	}()
	
	return nil
}
