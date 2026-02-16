// Package dto contains data transfer objects.
package dto

import "time"

// ReservedDto represents a room with its reservation slots
type ReservedDto struct {
	RoomID   int64             `json:"roomId"`
	RoomName string            `json:"roomName"`
	Slots    []ReservedSlotDto `json:"slots"`
}

// ReservedSlotDto represents a single reservation time slot.
type ReservedSlotDto struct {
	ID        int64     `json:"id"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	BookedBy  *string   `json:"bookedBy,omitempty"`
}

// ReservationDto is the returned dto after reservation creation.
type ReservationDto struct {
	ID        int64     `json:"Id"`
	RoomID    int64     `json:"roomId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	CreatedBy UserDto   `json:"createdBy"`
}

// UserDto is used in reservationDto to format createdby user
type UserDto struct {
	ID   int64  `json:"Id"`
	Name string `json:"name"`
}

// CreateReservationRequest is used to create reservation
type CreateReservationRequest struct {
	RoomID    int64     `json:"roomId" validate:"required,gt=0"`
	StartTime time.Time `json:"startTime" validate:"required,futureTime,schoolHours"`
	EndTime   time.Time `json:"endTime" validate:"required,gtfield=StartTime,schoolHours"`
}
