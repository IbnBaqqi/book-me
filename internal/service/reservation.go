// Package service contains business logic for the application.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/dto"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/google"
)

// User roles
const (
	RoleStudent = "STUDENT"
	RoleStaff   = "STAFF"
)

// ReservationService handles reservation business logic.
type ReservationService struct {
	db       *database.DB
	email    *email.Service
	calendar *google.CalendarService
}

// CreateReservationInput contains the input parameters for creating a reservation.
type CreateReservationInput struct {
	UserID    int64
	UserName  string
	UserRole  string
	RoomID    int64
	StartTime time.Time
	EndTime   time.Time
}

// GetReservationsInput contains the input parameters for fetching reservations.
type GetReservationsInput struct {
	StartDate time.Time
	EndDate   time.Time
	UserID    int64
	UserRole  string
}

// CancelReservationInput contains the input parameters for cancelling a reservation.
type CancelReservationInput struct {
	ID       int64
	UserID   int64
	UserRole string
}

// NewReservationService create dependencies for ReservationService.
func NewReservationService(
	db *database.DB,
	emailService *email.Service,
	calendarService *google.CalendarService,
) *ReservationService {
	return &ReservationService{
		db:       db,
		email:    emailService,
		calendar: calendarService,
	}
}

// CreateReservation is a service layer function that handles
// creating of reservation.
func (s *ReservationService) CreateReservation(
	ctx context.Context,
	input CreateReservationInput,
) (*database.Reservation, error) {

	// Get user email for sending email
	// TODO use redis instead
	dbUser, err := s.db.GetUser(ctx, input.UserID)
	if err != nil {
		slog.Error("failed to get user from db", "error", err)
		return nil, ErrGetUserFailed
	}

	// Fetch room
	room, err := s.db.GetRoomByID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	// Validate duration (students only)
	duration := input.EndTime.Sub(input.StartTime)
	maxDuration := 4 * time.Hour

	if duration > maxDuration && input.UserRole == RoleStudent {
		return nil, ErrExceedsMaxDuration
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("failed to start transaction: %v", err),
		}
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := s.db.WithTx(tx.Tx)

	// Check for overlapping reservations
	// Note: StartTime and EndTime are intentionally swapped for the overlap check logic
	// This checks if the new reservation's time range conflicts with existing ones
	overlap, err := qtx.ExistsOverlappingReservation(ctx, database.ExistsOverlappingReservationParams{
		RoomID:    input.RoomID,
		StartTime: input.EndTime,
		EndTime:   input.StartTime,
	})
	if err != nil {
		slog.Error("database error", "error", err)
		return nil, err
	}

	if overlap {
		return nil, ErrTimeSlotTaken
	}

	// Create reservation
	reservation, err := qtx.CreateReservation(ctx, database.CreateReservationParams{
		UserID:    input.UserID,
		RoomID:    room.ID,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Status:    "RESERVED",
	})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("failed to commit transaction: %v", err),
		}
	}

	// Create Google Calendar event
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()

		calendarReservation := &google.Reservation{
			StartTime: reservation.StartTime,
			EndTime:   reservation.EndTime,
			CreatedBy: input.UserName,
			Room:      room.Name,
		}

		eventID, err := s.calendar.CreateGoogleEvent(ctx, calendarReservation)
		if err != nil {
			slog.Error("Failed to create Google Calendar event", "error", err)
			return
		}

		// Update reservation with event ID
		if eventID != "" {
			updateErr := s.db.UpdateGoogleCalID(ctx, database.UpdateGoogleCalIDParams{
				ID:          reservation.ID,
				GcalEventID: sql.NullString{String: eventID, Valid: eventID != ""},
			})
			if updateErr != nil {
				slog.Warn("Failed to update reservation with calendar event ID", "error", updateErr)
			}
		}
	}()

	// Send confirmation email
	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()

		if err := s.email.SendConfirmation(
			emailCtx,
			dbUser.Email,
			room.Name,
			reservation.StartTime.Format("Monday, January 2, 2006 at 3:04 PM"),
			reservation.EndTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		); err != nil {
			slog.Error("failed to send confirmation email", "error", err)
		}
	}()

	return &reservation, nil
}

// GetReservations is a service layer function that handles
// fetching of reservation, grouping & formatting.
func (s *ReservationService) GetReservations(
	ctx context.Context,
	input GetReservationsInput,
) ([]dto.ReservedDto, error) {

	// Convert dates to datetime range
	startDateTime := input.StartDate
	endDateTime := input.EndDate.AddDate(0, 0, 1)

	isStaff := input.UserRole == RoleStaff

	// Fetch all reservations between dates
	reservations, err := s.db.GetAllBetweenDates(ctx, database.GetAllBetweenDatesParams{
		StartTime: startDateTime,
		EndTime:   endDateTime,
	})
	if err != nil {
		slog.Error("failed to fetch reservations from db", "error", err)
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
				StartTime: res.StartTime.UTC(),
				EndTime:   res.EndTime.UTC(),
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

// CancelReservation is a service layer function that handles
// cancelling of reservation.
func (s *ReservationService) CancelReservation(
	ctx context.Context,
	input CancelReservationInput,
) error {

	// Find reservation by ID
	reservation, err := s.db.GetReservationByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReservationNotFound
		}
		return err
	}

	isStaff := input.UserRole == RoleStaff
	isOwner := reservation.UserID == input.UserID

	if !isStaff && !isOwner {
		return ErrUnauthorizedCancellation
	}

	// Delete from database
	err = s.db.DeleteReservation(ctx, input.ID)
	if err != nil {
		return err
	}

	// Delete google calender event
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		if err := s.calendar.DeleteGoogleEvent(ctx, reservation.GcalEventID.String); err != nil {
			slog.Error("failed to delete google calendar event", "error", err)
		}
	}()

	return nil
}
