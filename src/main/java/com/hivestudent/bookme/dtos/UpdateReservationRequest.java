package com.hivestudent.bookme.dtos;

import jakarta.validation.constraints.NotNull;
import lombok.Data;
import org.springframework.format.annotation.DateTimeFormat;

import java.time.LocalDateTime;

@Data
public class UpdateReservationRequest {
    @NotNull(message = "Room id required")
    private Long roomId;

    @NotNull(message = "Start time required")
    @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME)
    private LocalDateTime startTime; // e.g. 06:00

    @NotNull(message = "End time required")
    @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME)
    private LocalDateTime endTime; // e.g. 07:00
}
