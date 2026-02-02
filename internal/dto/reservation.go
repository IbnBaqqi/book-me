package dto

import "time"

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